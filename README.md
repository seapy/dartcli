# dartcli

금융감독원 전자공시시스템(DART) 공시 정보를 터미널에서 조회하는 CLI 도구입니다.
기업 개황, 공시 목록, 재무정보, 공시 원문을 마크다운 형식으로 출력합니다.

사용하려면 [DART OpenAPI 인증키](https://opendart.fss.or.kr/uss/umt/EgovMberInsertView.do)가 필요합니다. 회원가입 후 즉시 발급받을 수 있어 어렵지 않습니다.

## 설치

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/seapy/dartcli/main/install.sh | sh
```

스크립트가 OS와 아키텍처를 자동으로 감지해서 `/usr/local/bin`에 설치합니다.
(`/usr/local/bin`에 쓰기 권한이 없으면 자동으로 `sudo`를 사용합니다.)

### Windows

[Releases 페이지](https://github.com/seapy/dartcli/releases/latest)에서 `dartcli_windows_amd64.zip`을 다운로드한 뒤, `dartcli.exe`를 PATH에 포함된 폴더(예: `C:\Windows\System32` 또는 별도 지정 폴더)에 복사합니다.

### 설치 확인

```bash
dartcli version
```

## API 키 설정

[DART OpenAPI](https://opendart.fss.or.kr/uss/umt/EgovMberInsertView.do)에서 무료로 발급받을 수 있습니다.

> **[인증키 발급 소요시간]**
> - 개인회원: 계정신청 완료 후 즉시 발급
> - 기업회원: 계정신청 완료 및 담당자 승인 (1~2영업일) 후 발급

```bash
# 환경변수 (권장)
export DART_API_KEY=<발급받은_키>

# 설정파일에 영구 저장
echo "api_key: <발급받은_키>" > ~/.dartcli/config.yaml
```

`search`를 제외한 모든 명령에서 API 키가 필요합니다. (`search`는 로컬 캐시만 사용)

---

## 명령어

### `search` — 기업 검색

기업명이나 종목코드로 Corp Code를 조회합니다. 다른 명령에 입력할 정확한 이름을 확인할 때 사용합니다.

```bash
dartcli search 삼성
dartcli search 005930
```

```
# 검색 결과: "삼성"

 기업명   | 종목코드 | Corp Code | 수정일
----------|----------|-----------|----------
 삼성     | -        | 00893765  | 20230213
```

---

### `company` — 기업 개황

대표이사, 설립일, 주소, 홈페이지 등 기업 기본 정보를 조회합니다.

```bash
dartcli company 삼성전자
dartcli company 005930        # 종목코드로도 조회 가능
```

```
# 삼성전자(주) (005930)

 항목        | 내용
-------------|-----------------------------
 법인구분    | 유가증권시장
 대표이사    | 전영현, 노태문
 설립일      | 1969-01-13
 결산월      | 12월
 주소        | 경기도 수원시 영통구 삼성로 129
 홈페이지    | www.samsung.com
 전화번호    | 02-2255-0114
```

---

### `list` — 공시 목록

최근 공시 목록을 조회합니다. 출력된 **접수번호**를 `view` 명령에 사용합니다.

```bash
dartcli list 삼성전자                          # 최근 5년 공시 (기본 20건)
dartcli list 삼성전자 --type A --limit 5       # 정기공시만 최대 5건
dartcli list 삼성전자 --days 90                # 최근 90일
dartcli list 삼성전자 --start 20240101 --end 20241231
```

**`--type` 코드:**

| 코드 | 종류 |
|------|------|
| A | 정기공시 (사업·분기·반기보고서) |
| B | 주요사항보고 |
| C | 발행공시 |
| D | 지분공시 |
| E | 기타공시 |
| F | 외부감사관련 |

```
 접수일     | 공시명                    | 제출인   | 접수번호
------------|---------------------------|----------|------------------
 2025-11-14 | 분기보고서 (2025.09)      | 삼성전자 | 20251114002447
 2025-08-14 | 반기보고서 (2025.06)      | 삼성전자 | 20250814003156
 2025-03-11 | 사업보고서 (2024.12) [연] | 삼성전자 | 20250311001085
```

---

### `finance` — 재무정보

재무상태표·손익계산서·현금흐름표를 억원 단위로 조회합니다. 전기 대비 증감률도 함께 표시됩니다.

```bash
dartcli finance 삼성전자                       # 작년 연간 연결 재무제표 (기본값)
dartcli finance 삼성전자 --year 2024
dartcli finance 삼성전자 --year 2024 --period half   # 반기
dartcli finance 삼성전자 --year 2024 --type ofs       # 개별 재무제표
```

**`--period` 옵션:** `annual`(연간, 기본) | `q1`(1분기) | `half`(반기) | `q3`(3분기)
**`--type` 옵션:** `cfs`(연결, 기본) | `ofs`(개별)

```
# 삼성전자 재무정보
**2024년 연결 연간 기준**

## 손익계산서

 계정과목           | 당기          | 전기          | 증감률
--------------------|---------------|---------------|--------
 매출액             | 3008709.0억   | 2589354.9억   | +16.2%
 영업이익           | 327259.6억    | 65669.8억     | +398.3%
 당기순이익(손실)   | 344513.5억    | 154871.0억    | +122.5%
```

---

### `view` — 공시 원문 조회

`list`에서 확인한 **접수번호**로 공시 원문을 터미널에서 바로 읽을 수 있습니다.
사업보고서, 분기·반기보고서, 감사보고서 등 모든 공시 유형을 지원합니다.

```bash
dartcli view 20251114002447                    # 터미널에서 마크다운으로 렌더링 (기본)
dartcli view 20251114002447 --browser          # DART 웹사이트에서 브라우저로 열기
dartcli view 20251114002447 --download         # ZIP 원문 파일로 저장 (접수번호.zip)
dartcli view 20251114002447 --download -o ./samsung_q3.zip
```

**출력 구조:**

DART XML 원문을 파싱해서 다음과 같은 마크다운 구조로 변환합니다.

```
# 분기보고서                         ← 문서 제목
| 회사명 | 사업연도 | 제출일 | …      ← 표지 정보 (표)
## 목 차                             ← 목차 (표)
## I. 회사의 개요                    ← 본문 섹션 (## 헤딩)
### 1. 회사의 개요                   ← 하위 섹션 (### 헤딩)
   본문 서술 텍스트 …
| 항목 | 당기 | 전기 | …            ← 재무제표·통계 (마크다운 표)
## II. 사업의 내용
…
```

사업보고서 기준 수천 줄 분량입니다. 파이프로 연결하면 색상이 자동으로 비활성화됩니다:

```bash
dartcli view 20251114002447 | less -R
dartcli view 20251114002447 | grep -A 30 "사업의 개요"
```

---

### `cache` — 캐시 관리

전국 약 12,000개 기업의 Corp Code를 로컬에 캐싱합니다. 7일이 지나면 자동으로 갱신됩니다.

```bash
dartcli cache status    # 캐시 파일 경로 및 최종 갱신 시각 확인
dartcli cache refresh   # 즉시 갱신
dartcli cache clear     # 캐시 삭제
```

캐시 파일 위치: `~/.dartcli/cache/corpcode.json`

---

## 전역 옵션

모든 명령에 사용할 수 있습니다.

| 옵션 | 설명 |
|------|------|
| `--api-key <키>` | DART API 키 (환경변수·설정파일보다 우선) |
| `--no-color` | 색상 출력 비활성화 |
| `--style <스타일>` | 렌더링 스타일: `auto`(기본) \| `dark` \| `light` \| `notty` |
| `--config <경로>` | 설정파일 경로 (기본: `~/.dartcli/config.yaml`) |

---

## 사용 예시 (워크플로)

```bash
# 1. 기업 검색으로 정확한 이름 확인
dartcli search 카카오

# 2. 기업 개황 조회
dartcli company 카카오

# 3. 최근 정기공시 목록 확인
dartcli list 카카오 --type A --limit 10

# 4. 가장 최근 사업보고서 원문 읽기 (접수번호는 list 결과에서 복사)
dartcli view 20250328001234

# 5. 재무정보 요약 조회
dartcli finance 카카오 --year 2024
```

---

## AI 에이전트 활용 가이드

### 에이전트에게 설치 요청하기

AI 에이전트(Claude, Cursor 등)에게 설치를 맡길 때는 **반드시 전체 저장소 경로**를 명시하세요. `dartcli`라고만 하면 엉뚱한 패키지를 설치할 수 있습니다.

```
github.com/seapy/dartcli 설치해서 삼성전자 공시정보 찾아봐
```

구조화된 마크다운 출력이므로 LLM 컨텍스트에 직접 삽입하기에 적합합니다.

```bash
# 색상 코드 없이 순수 텍스트 출력
dartcli finance 삼성전자 --no-color
dartcli view 20251114002447 --no-color

# 또는 파이프를 통해 자동으로 색상 제거
dartcli view 20251114002447 | cat
```

**기업명이 불명확한 경우**: 동명 기업이 여러 개일 때 인터랙티브 선택 프롬프트가 나타납니다. 에이전트 환경처럼 stdin이 없는 경우 `search`로 먼저 Corp Code를 확인하고, 정확한 기업명(또는 종목코드)을 사용하세요.

```bash
# 검색으로 정확한 이름 확인 후 사용
dartcli search 카카오
dartcli company "카카오(주)"    # 검색 결과의 기업명을 그대로 사용
```

**`view` 출력 규모**: 삼성전자 사업보고서 기준 약 15,000줄 이상입니다. 특정 섹션만 필요하다면 grep으로 필터링하세요.

```bash
dartcli view 20250311001085 | grep -A 30 "사업의 개요"
```
