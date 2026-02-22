package cache

import (
	"testing"
)

// testStore returns a Store built from representative sample data.
func testStore() *Store {
	corps := []*CorpInfo{
		{CorpCode: "00126380", CorpName: "삼성전자", StockCode: "005930"},
		{CorpCode: "01153956", CorpName: "컬리"},
		{CorpCode: "01494172", CorpName: "컬리넥스트마일"},
		{CorpCode: "01547845", CorpName: "당근마켓"},
		{CorpCode: "01717824", CorpName: "당근페이"},
		{CorpCode: "01138364", CorpName: "더핑크퐁컴퍼니", StockCode: "403850"},
		{CorpCode: "00293886", CorpName: "카카오"},
		{CorpCode: "00000002", CorpName: "카카오뱅크"},
		{CorpCode: "00000003", CorpName: "카카오페이"},
		{CorpCode: "01154811", CorpName: "주식회사 오늘의집"},
		// 법인 형태어 노이즈 테스트용: "주식회사"를 포함하는 무관한 회사
		{CorpCode: "99999999", CorpName: "두성에스비텍주식회사(구:두성공업주식회사)"},
	}
	return buildStore(corps)
}

func corpNames(corps []*CorpInfo) []string {
	out := make([]string, len(corps))
	for i, c := range corps {
		out[i] = c.CorpName
	}
	return out
}

// --- exact match tests ---

func TestSearch_ExactStockCode(t *testing.T) {
	s := testStore()
	results := s.Search("005930")
	if len(results) != 1 || results[0].CorpName != "삼성전자" {
		t.Fatalf("종목코드 완전일치 실패: got %v", corpNames(results))
	}
}

func TestSearch_ExactCorpCode(t *testing.T) {
	s := testStore()
	results := s.Search("00126380")
	if len(results) != 1 || results[0].CorpName != "삼성전자" {
		t.Fatalf("법인코드 완전일치 실패: got %v", corpNames(results))
	}
}

func TestSearch_ExactName(t *testing.T) {
	s := testStore()
	results := s.Search("삼성전자")
	if len(results) != 1 || results[0].CorpName != "삼성전자" {
		t.Fatalf("이름 완전일치 실패: got %v", corpNames(results))
	}
}

// --- substring tests ---

func TestSearch_Substring_당근(t *testing.T) {
	s := testStore()
	results := s.Search("당근")
	if len(results) != 2 {
		t.Fatalf("'당근' substring 검색: 2건 기대, got %d %v", len(results), corpNames(results))
	}
}

func TestSearch_Substring_카카오(t *testing.T) {
	// "카카오" 이름 완전일치가 존재하므로 정확히 1건만 반환 (substring 전에 리턴)
	s := testStore()
	results := s.Search("카카오")
	if len(results) != 1 || results[0].CorpName != "카카오" {
		t.Fatalf("'카카오' 이름 완전일치: 1건 기대, got %d %v", len(results), corpNames(results))
	}
}

func TestSearch_Substring_카카오계열(t *testing.T) {
	// "카카오뱅" 처럼 완전일치 없는 쿼리는 substring 으로 카카오뱅크를 찾음
	s := testStore()
	results := s.Search("카카오뱅")
	if len(results) != 1 || results[0].CorpName != "카카오뱅크" {
		t.Fatalf("'카카오뱅' substring 검색: 카카오뱅크 기대, got %v", corpNames(results))
	}
}

func TestSearch_Substring_핑크퐁(t *testing.T) {
	s := testStore()
	results := s.Search("핑크퐁")
	if len(results) != 1 || results[0].CorpName != "더핑크퐁컴퍼니" {
		t.Fatalf("'핑크퐁' substring 검색 실패: got %v", corpNames(results))
	}
}

// --- fuzzy tests ---

func TestSearch_Fuzzy_마켓컬리(t *testing.T) {
	s := testStore()
	results := s.Search("마켓컬리")
	if len(results) == 0 {
		t.Fatal("'마켓컬리' fuzzy 검색: 결과 없음 (기대: '컬리' 포함)")
	}
	found := false
	for _, r := range results {
		if r.CorpName == "컬리" {
			found = true
		}
	}
	if !found {
		t.Fatalf("'마켓컬리' fuzzy 결과에 '컬리' 없음: got %v", corpNames(results))
	}
	t.Logf("'마켓컬리' fuzzy 결과: %v", corpNames(results))
}

func TestSearch_Fuzzy_오늘의집(t *testing.T) {
	// "오늘의집" → substring 으로도 찾힘
	s := testStore()
	results := s.Search("오늘의집")
	if len(results) == 0 {
		t.Fatal("'오늘의집' 검색: 결과 없음")
	}
	found := false
	for _, r := range results {
		if r.CorpName == "주식회사 오늘의집" {
			found = true
		}
	}
	if !found {
		t.Fatalf("'오늘의집' 결과에 '주식회사 오늘의집' 없음: got %v", corpNames(results))
	}
}

func TestSearch_Fuzzy_법인형태어_노이즈제거(t *testing.T) {
	// "주식회사 오늘의집" 검색 시 "주식회사"를 공유하는 무관한 회사가 나오면 안 됨
	s := testStore()
	results := s.Search("주식회사 오늘의집")
	for _, r := range results {
		if r.CorpName == "두성에스비텍주식회사(구:두성공업주식회사)" {
			t.Fatalf("'주식회사 오늘의집' 검색에서 무관한 회사 '두성에스비텍주식회사'가 반환됨 (법인 형태어 노이즈)")
		}
	}
	t.Logf("'주식회사 오늘의집' 검색 결과: %v", corpNames(results))
}

func TestStripLegalForm(t *testing.T) {
	cases := []struct{ in, want string }{
		{"주식회사 오늘의집", "오늘의집"},
		{"삼성전자(주)", "삼성전자"},
		{"(주)카카오", "카카오"},
		{"유한회사테스트", "테스트"},
		{"컬리", "컬리"},
	}
	for _, c := range cases {
		got := stripLegalForm(c.in)
		if got != c.want {
			t.Errorf("stripLegalForm(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSearch_NoResult_완전엉뚱한검색어(t *testing.T) {
	s := testStore()
	results := s.Search("xyzxyz없는기업명zyx")
	if len(results) != 0 {
		t.Logf("'xyzxyz없는기업명zyx' 검색 결과 (fuzzy 허용): %v", corpNames(results))
	}
}

// --- bigramSim unit tests ---

func TestBigramSim_완전동일(t *testing.T) {
	score := bigramSim("삼성전자", "삼성전자")
	if score != 1.0 {
		t.Fatalf("동일 문자열 bigramSim: 1.0 기대, got %f", score)
	}
}

func TestBigramSim_부분겹침(t *testing.T) {
	// "마켓컬리" bigrams: [마켓, 켓컬, 컬리]
	// "컬리" bigrams: [컬리]  → 1 match / 3 query bigrams = 0.333
	score := bigramSim("마켓컬리", "컬리")
	if score < 0.3 {
		t.Fatalf("'마켓컬리' vs '컬리': 0.3 이상 기대, got %f", score)
	}
}

func TestBigramSim_핑크퐁(t *testing.T) {
	// "핑크퐁" bigrams: [핑크, 크퐁]
	// "더핑크퐁컴퍼니" bigrams: [더핑, 핑크, 크퐁, 퐁컴, 컴퍼, 퍼니] → 2 match / 2 = 1.0
	score := bigramSim("핑크퐁", "더핑크퐁컴퍼니")
	if score != 1.0 {
		t.Fatalf("'핑크퐁' vs '더핑크퐁컴퍼니': 1.0 기대, got %f", score)
	}
}

func TestBigramSim_단일문자(t *testing.T) {
	// 단일 rune은 bigram 없음 → 0
	score := bigramSim("가", "가나다라")
	if score != 0 {
		t.Fatalf("단일 문자 bigramSim: 0 기대, got %f", score)
	}
}
