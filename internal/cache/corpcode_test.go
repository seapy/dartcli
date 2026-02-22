package cache

import (
	"os"
	"testing"
)

// testStore builds a Store from representative sample data.
func testStore() *Store {
	corps := []*CorpInfo{
		{CorpCode: "00126380", CorpName: "삼성전자", StockCode: "005930"},
		{CorpCode: "01153956", CorpName: "컬리"},
		{CorpCode: "01494172", CorpName: "컬리넥스트마일"},
		{CorpCode: "01713402", CorpName: "컬리페이"},
		{CorpCode: "01547845", CorpName: "당근마켓"},
		{CorpCode: "01717824", CorpName: "당근페이"},
		{CorpCode: "01138364", CorpName: "더핑크퐁컴퍼니", StockCode: "403850"},
		{CorpCode: "00293886", CorpName: "카카오"},
		{CorpCode: "00000002", CorpName: "카카오뱅크"},
		{CorpCode: "00000003", CorpName: "카카오페이"},
		{CorpCode: "01154811", CorpName: "주식회사 오늘의집"},
		// "마켓컬리" 검색 시 노이즈가 될 수 있는 실제 데이터와 유사한 회사들
		{CorpCode: "N001", CorpName: "게이트마켓"},
		{CorpCode: "N002", CorpName: "지마켓"},
		{CorpCode: "N003", CorpName: "알루마켓"},
		{CorpCode: "N004", CorpName: "비즈마켓"},
		{CorpCode: "N005", CorpName: "올인마켓"},
		{CorpCode: "N006", CorpName: "레어마켓"},
		{CorpCode: "N007", CorpName: "마켓비"},
		{CorpCode: "N008", CorpName: "와마켓"},
		{CorpCode: "N009", CorpName: "맥쿼리IMM마켓뉴트럴혼합형사모펀드"},
		{CorpCode: "N010", CorpName: "다이와증권캐피탈마켓서울지점"},
		{CorpCode: "N011", CorpName: "에이치앤디마켓플레이스"},
		{CorpCode: "N012", CorpName: "코리아마켓팅"},
		// "주식회사 오늘의집" 검색 시 법인 형태어 노이즈
		{CorpCode: "N099", CorpName: "두성에스비텍주식회사(구:두성공업주식회사)"},
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
	// "카카오" 이름 완전일치가 존재하므로 1건만 반환 (substring 전에 리턴)
	s := testStore()
	results := s.Search("카카오")
	if len(results) != 1 || results[0].CorpName != "카카오" {
		t.Fatalf("'카카오' 이름 완전일치: 1건 기대, got %d %v", len(results), corpNames(results))
	}
}

func TestSearch_Substring_카카오계열(t *testing.T) {
	// 완전일치 없는 쿼리는 substring으로 찾음
	s := testStore()
	results := s.Search("카카오뱅")
	if len(results) != 1 || results[0].CorpName != "카카오뱅크" {
		t.Fatalf("'카카오뱅' substring 검색 실패: got %v", corpNames(results))
	}
}

func TestSearch_Substring_핑크퐁(t *testing.T) {
	s := testStore()
	results := s.Search("핑크퐁")
	if len(results) != 1 || results[0].CorpName != "더핑크퐁컴퍼니" {
		t.Fatalf("'핑크퐁' substring 검색 실패: got %v", corpNames(results))
	}
}

func TestSearch_Substring_오늘의집(t *testing.T) {
	// "오늘의집"은 "주식회사 오늘의집"의 substring
	s := testStore()
	results := s.Search("오늘의집")
	if len(results) != 1 || results[0].CorpName != "주식회사 오늘의집" {
		t.Fatalf("'오늘의집' substring 검색 실패: got %v", corpNames(results))
	}
}

// --- fuzzy tests ---

// TestSearch_Fuzzy_마켓컬리_컬리상위랭크 는 핵심 케이스:
// "마켓컬리" 검색 시 마켓XX 회사들이 다수 있어도 "컬리"가 1위여야 한다.
// (Dice 계수 기반: 컬리=0.5, 마켓4글자회사=0.333)
func TestSearch_Fuzzy_마켓컬리_컬리상위랭크(t *testing.T) {
	s := testStore()
	results := s.Search("마켓컬리")
	if len(results) == 0 {
		t.Fatal("'마켓컬리' fuzzy 검색: 결과 없음")
	}
	if results[0].CorpName != "컬리" {
		t.Fatalf("'마켓컬리' fuzzy 1위는 '컬리'여야 함: got %v", corpNames(results))
	}
	t.Logf("'마켓컬리' fuzzy 결과: %v", corpNames(results))
}

// TestSearch_Fuzzy_법인형태어_노이즈제거: "주식회사"만 공유하는 무관한 회사가 나오면 안 됨
func TestSearch_Fuzzy_법인형태어_노이즈제거(t *testing.T) {
	s := testStore()
	results := s.Search("주식회사 오늘의집")
	for _, r := range results {
		if r.CorpName == "두성에스비텍주식회사(구:두성공업주식회사)" {
			t.Fatalf("'주식회사 오늘의집' 검색에서 무관한 회사 '두성에스비텍주식회사'가 반환됨 (법인 형태어 노이즈)")
		}
	}
	t.Logf("'주식회사 오늘의집' 검색 결과: %v", corpNames(results))
}

func TestSearch_NoResult_완전엉뚱한검색어(t *testing.T) {
	s := testStore()
	results := s.Search("xyzxyz없는기업명zyx")
	// 완전히 무관한 검색어는 결과가 없어야 함 (있어도 오탐이지만 경고만)
	if len(results) != 0 {
		t.Logf("경고: 무관한 검색어에 fuzzy 결과 %d건: %v", len(results), corpNames(results))
	}
}

// --- bigramSim unit tests ---

func TestBigramSim_완전동일(t *testing.T) {
	// 동일 문자열은 1.0
	score := bigramSim("삼성전자", "삼성전자")
	if score != 1.0 {
		t.Fatalf("동일 문자열 bigramSim: 1.0 기대, got %f", score)
	}
}

func TestBigramSim_Dice_짧은타겟이_높은점수(t *testing.T) {
	// Dice 계수 검증: 짧은 타겟("컬리")이 긴 타겟("컬리넥스트마일")보다 높은 점수여야 함
	// "마켓컬리" (3 bigrams) vs "컬리" (1 bigram): 2*1/(3+1)=0.500
	// "마켓컬리" (3 bigrams) vs "컬리넥스트마일" (6 bigrams): 2*1/(3+6)=0.222
	short := bigramSim("마켓컬리", "컬리")
	long := bigramSim("마켓컬리", "컬리넥스트마일")
	if short <= long {
		t.Fatalf("'컬리' score(%f) > '컬리넥스트마일' score(%f) 이어야 함", short, long)
	}
	if short < 0.3 {
		t.Fatalf("'마켓컬리' vs '컬리': threshold 0.3 이상 기대, got %f", short)
	}
}

func TestBigramSim_긴노이즈_필터링(t *testing.T) {
	// 긴 회사명은 Dice로 낮아져서 threshold 이하가 되어야 함
	// "마켓컬리" vs "맥쿼리IMM마켓뉴트럴혼합형사모펀드": 겹치는 bigram 1개, 타겟이 매우 길어 Dice<0.3
	score := bigramSim("마켓컬리", "맥쿼리imm마켓뉴트럴혼합형사모펀드")
	if score >= 0.3 {
		t.Fatalf("긴 노이즈 회사명은 threshold 0.3 미만이어야 함, got %f", score)
	}
}

func TestBigramSim_단일문자(t *testing.T) {
	// 단일 rune은 bigram 없음 → 0
	score := bigramSim("가", "가나다라")
	if score != 0 {
		t.Fatalf("단일 문자 bigramSim: 0 기대, got %f", score)
	}
}

// --- stripLegalForm tests ---

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

// --- integration tests (실제 캐시 데이터 기반) ---

// realStore 는 실제 캐시가 없으면 nil 반환.
func realStore(t *testing.T) *Store {
	t.Helper()
	path, err := CorpCodePath()
	if err != nil {
		t.Skip("캐시 경로 확인 불가:", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("캐시 파일 없음 — dartcli cache refresh 후 재시도")
	}
	store, err := loadFromDisk(path)
	if err != nil {
		t.Skip("캐시 로드 실패:", err)
	}
	return store
}

func TestIntegration_마켓컬리_컬리반환(t *testing.T) {
	s := realStore(t)
	results := s.Search("마켓컬리")
	if len(results) == 0 {
		t.Fatal("'마켓컬리' 검색: 결과 없음 (기대: 컬리 포함)")
	}
	// "컬리"가 결과에 포함되어야 함
	found := false
	for _, r := range results {
		if r.CorpName == "컬리" {
			found = true
		}
	}
	if !found {
		t.Fatalf("'마켓컬리' 결과에 '컬리' 없음: got %v", corpNames(results))
	}
	// "컬리"가 1등이어야 함 (Dice 계수로 짧은 매칭이 우선)
	if results[0].CorpName != "컬리" {
		t.Errorf("'마켓컬리' 1위 기대='컬리', got='%s' (전체: %v)", results[0].CorpName, corpNames(results))
	}
	t.Logf("결과: %v", corpNames(results))
}

func TestIntegration_삼성전자(t *testing.T) {
	s := realStore(t)
	results := s.Search("삼성전자")
	if len(results) != 1 || results[0].CorpName != "삼성전자" {
		t.Fatalf("'삼성전자' 검색 실패: got %v", corpNames(results))
	}
}

func TestIntegration_당근(t *testing.T) {
	s := realStore(t)
	results := s.Search("당근")
	names := corpNames(results)
	foundMarket := false
	for _, n := range names {
		if n == "당근마켓" {
			foundMarket = true
		}
	}
	if !foundMarket {
		t.Fatalf("'당근' 검색에 '당근마켓' 없음: got %v", names)
	}
}

func TestIntegration_핑크퐁(t *testing.T) {
	s := realStore(t)
	results := s.Search("핑크퐁")
	if len(results) == 0 {
		t.Fatal("'핑크퐁' 검색: 결과 없음")
	}
	if results[0].CorpName != "더핑크퐁컴퍼니" {
		t.Fatalf("'핑크퐁' 1위 기대='더핑크퐁컴퍼니', got='%s'", results[0].CorpName)
	}
}

func TestIntegration_종목코드_삼성전자(t *testing.T) {
	s := realStore(t)
	results := s.Search("005930")
	if len(results) != 1 || results[0].CorpName != "삼성전자" {
		t.Fatalf("종목코드 '005930' 검색 실패: got %v", corpNames(results))
	}
}
