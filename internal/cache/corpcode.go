package cache

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/seapy/dartcli/internal/httpclient"
)

const (
	corpCodeURL   = "https://opendart.fss.or.kr/api/corpCode.xml"
	cacheMaxAge   = 7 * 24 * time.Hour
)

// corpCodeXML is the XML structure of CORPCODE.xml inside the ZIP.
type corpCodeXML struct {
	XMLName xml.Name      `xml:"result"`
	List    []corpCodeRow `xml:"list"`
}

type corpCodeRow struct {
	CorpCode   string `xml:"corp_code"`
	CorpName   string `xml:"corp_name"`
	StockCode  string `xml:"stock_code"`
	ModifyDate string `xml:"modify_date"`
}

// Store is the in-memory index of corp codes.
type Store struct {
	byCode map[string]*CorpInfo // corp_code -> CorpInfo
	byName map[string][]*CorpInfo // corp_name (lower) -> []CorpInfo
	byStock map[string]*CorpInfo // stock_code -> CorpInfo
	All    []*CorpInfo
}

// Refresh downloads and rebuilds the corp code cache.
func Refresh(apiKey string) (*Store, error) {
	data, err := downloadCorpCodeZIP(apiKey)
	if err != nil {
		return nil, err
	}

	corps, err := parseCorpCodeZIP(data)
	if err != nil {
		return nil, err
	}

	path, err := CorpCodePath()
	if err != nil {
		return nil, err
	}

	if err := saveCorpCodeJSON(corps, path); err != nil {
		return nil, err
	}

	return buildStore(corps), nil
}

// Load loads the corp code cache from disk (auto-refreshes if stale).
// Returns the store and whether a refresh occurred.
func Load(apiKey string) (*Store, bool, error) {
	path, err := CorpCodePath()
	if err != nil {
		return nil, false, err
	}

	refreshed := false
	if needsRefresh(path) {
		store, err := Refresh(apiKey)
		if err != nil {
			// Try loading stale cache
			store, loadErr := loadFromDisk(path)
			if loadErr != nil {
				return nil, false, fmt.Errorf("cache refresh failed: %w; load fallback failed: %v", err, loadErr)
			}
			fmt.Fprintf(os.Stderr, "warning: cache refresh failed (%v), using stale cache\n", err)
			return store, false, nil
		}
		return store, true, nil
	}

	store, err := loadFromDisk(path)
	if err != nil {
		// Cache is missing or corrupt, try refreshing
		store, err = Refresh(apiKey)
		if err != nil {
			return nil, false, err
		}
		refreshed = true
	}
	return store, refreshed, nil
}

// Search finds corporations matching the query.
// Priority: exact stock/corp code → exact name → substring → bigram fuzzy.
func (s *Store) Search(query string) []*CorpInfo {
	// Try stock code (exact)
	if c, ok := s.byStock[query]; ok {
		return []*CorpInfo{c}
	}
	// Try corp code (exact)
	if c, ok := s.byCode[query]; ok {
		return []*CorpInfo{c}
	}
	// Exact name match
	lower := strings.ToLower(query)
	if results, ok := s.byName[lower]; ok {
		return results
	}
	// Substring search
	var matches []*CorpInfo
	for _, info := range s.All {
		if strings.Contains(strings.ToLower(info.CorpName), lower) {
			matches = append(matches, info)
		}
	}
	if len(matches) > 0 {
		return matches
	}
	// Fuzzy fallback: bigram similarity
	return s.fuzzySearch(lower, 0.3, 10)
}

// fuzzySearch returns up to max corps whose name has bigram similarity ≥ threshold
// with query, sorted by score descending.
func (s *Store) fuzzySearch(query string, threshold float64, max int) []*CorpInfo {
	type scored struct {
		info  *CorpInfo
		score float64
	}
	// Strip legal form words from the query once before the loop.
	normalizedQuery := stripLegalForm(query)
	if normalizedQuery == "" {
		return nil
	}

	var results []scored
	for _, info := range s.All {
		normalizedTarget := stripLegalForm(strings.ToLower(info.CorpName))
		score := bigramSim(normalizedQuery, normalizedTarget)
		if score >= threshold {
			results = append(results, scored{info, score})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	if len(results) > max {
		results = results[:max]
	}
	out := make([]*CorpInfo, len(results))
	for i, r := range results {
		out[i] = r.info
	}
	return out
}

// legalFormReplacer strips common Korean legal entity form words so they don't
// inflate bigram similarity scores (e.g. "주식회사" appears in thousands of names).
var legalFormReplacer = strings.NewReplacer(
	"주식회사", "",
	"유한회사", "",
	"합자회사", "",
	"합명회사", "",
	"유한책임회사", "",
	"(주)", "",
	"(유)", "",
)

func stripLegalForm(s string) string {
	return strings.TrimSpace(legalFormReplacer.Replace(s))
}

// bigramSim returns the fraction of query's character bigrams that appear in target.
// Uses rune-level bigrams for correct Korean handling.
func bigramSim(query, target string) float64 {
	qr := []rune(query)
	tr := []rune(target)
	if len(qr) < 2 {
		return 0
	}
	// Build target bigram frequency map
	tBig := make(map[[2]rune]int, len(tr))
	for i := 0; i < len(tr)-1; i++ {
		tBig[[2]rune{tr[i], tr[i+1]}]++
	}
	// Count how many query bigrams appear in target
	matched := 0
	for i := 0; i < len(qr)-1; i++ {
		k := [2]rune{qr[i], qr[i+1]}
		if tBig[k] > 0 {
			matched++
			tBig[k]--
		}
	}
	return float64(matched) / float64(len(qr)-1)
}

// Status returns cache file info.
func Status() (exists bool, modTime time.Time, stale bool, err error) {
	path, err := CorpCodePath()
	if err != nil {
		return false, time.Time{}, false, err
	}
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, time.Time{}, true, nil
	}
	if err != nil {
		return false, time.Time{}, false, err
	}
	mt := fi.ModTime()
	return true, mt, time.Since(mt) > cacheMaxAge, nil
}

// Clear removes the cache file.
func Clear() error {
	path, err := CorpCodePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// --- internal helpers ---

func needsRefresh(path string) bool {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return true
	}
	if err != nil {
		return true
	}
	return time.Since(fi.ModTime()) > cacheMaxAge
}

func downloadCorpCodeZIP(apiKey string) ([]byte, error) {
	params := url.Values{}
	params.Set("crtfc_key", apiKey)
	u := corpCodeURL + "?" + params.Encode()

	client := httpclient.New(60 * time.Second)
	resp, err := client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("downloading corp code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d downloading corp code", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func parseCorpCodeZIP(data []byte) ([]*CorpInfo, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("reading ZIP: %w", err)
	}

	for _, f := range zr.File {
		if !strings.EqualFold(f.Name, "CORPCODE.xml") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening CORPCODE.xml: %w", err)
		}
		defer rc.Close()

		xmlData, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("reading CORPCODE.xml: %w", err)
		}

		var root corpCodeXML
		if err := xml.Unmarshal(xmlData, &root); err != nil {
			return nil, fmt.Errorf("parsing CORPCODE.xml: %w", err)
		}

		corps := make([]*CorpInfo, 0, len(root.List))
		for _, row := range root.List {
			corps = append(corps, &CorpInfo{
				CorpCode:   strings.TrimSpace(row.CorpCode),
				CorpName:   strings.TrimSpace(row.CorpName),
				StockCode:  strings.TrimSpace(row.StockCode),
				ModifyDate: strings.TrimSpace(row.ModifyDate),
			})
		}
		return corps, nil
	}

	return nil, fmt.Errorf("CORPCODE.xml not found in ZIP")
}

func saveCorpCodeJSON(corps []*CorpInfo, path string) error {
	data, err := json.Marshal(corps)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func loadFromDisk(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var corps []*CorpInfo
	if err := json.Unmarshal(data, &corps); err != nil {
		return nil, err
	}
	return buildStore(corps), nil
}

func buildStore(corps []*CorpInfo) *Store {
	s := &Store{
		byCode:  make(map[string]*CorpInfo, len(corps)),
		byName:  make(map[string][]*CorpInfo),
		byStock: make(map[string]*CorpInfo),
		All:     corps,
	}
	for _, c := range corps {
		s.byCode[c.CorpCode] = c
		key := strings.ToLower(c.CorpName)
		s.byName[key] = append(s.byName[key], c)
		if c.StockCode != "" {
			s.byStock[c.StockCode] = c
		}
	}
	return s
}
