# Audit SEO / ASO — search engine + app store optimization

## 1. 목적

**SEO** (Search Engine Optimization, web) + **ASO** (App Store Optimization, mobile) 의 baseline audit + 개선 우선순위. *organic acquisition* 의 main lever.

`decide-form-factor-app-vs-web` 산출에 따라 SEO / ASO / 둘 다 적용.

## 2. 사용 시점

- §8 iterate-product 의 *분기 organic 분석*
- 신규 web / mobile launch 후 — baseline audit
- ranking drop 시 — drift 식별
- 경쟁사가 ranking 추월 시
- 신규 keyword / category 진입 검토

## 3. 입력

### 필수
- `decide-form-factor-app-vs-web` 산출 — web / mobile / both
- 현재 site / app store listing
- target keyword 가설 list

### 선택
- 경쟁사 ranking (Ahrefs / Semrush / Sensor Tower)
- 사용자 검색 데이터 (자체 search log)

## 4. Stage 흐름

### Stage 1: SEO audit (web)

| 영역 | 검증 |
|------|-----|
| Technical | Core Web Vitals (LCP / FID / CLS), mobile-friendly, HTTPS, sitemap, robots.txt |
| On-page | title tag (60 char) / meta description (160 char) / heading hierarchy / internal link |
| Content | keyword 정합 / 길이 / 신선도 / E-E-A-T (experience, expertise, authority, trust) |
| Off-page | backlink quality + diversity, brand mention |
| Schema | structured data (Article / Product / FAQ / HowTo) |

도구: Google Search Console / Lighthouse / Ahrefs / Semrush / Screaming Frog.

### Stage 2: ASO audit (mobile)

| 영역 | 검증 |
|------|-----|
| Metadata | title (30 char) / subtitle (30 char) / keyword field (100 char iOS) |
| Visual | icon / screenshot (5+) / preview video |
| Description | first 3 줄 critical (above the fold) |
| Rating | 4.5+ 목표 / review response 빈도 |
| Category | primary + secondary 적절성 |
| Localization | locale 별 별도 metadata + screenshot |

도구: Sensor Tower / data.ai / App Annie.

### Stage 3: Keyword research

| 차원 | 측정 |
|------|------|
| Search volume | monthly searches |
| Difficulty | competitor strength |
| Intent | informational / navigational / transactional / commercial |
| Trend | rising / stable / declining |

→ low difficulty + high intent (commercial) 우선. high difficulty 는 long-tail variant 로.

### Stage 4: Content gap 분석

경쟁사 vs 우리:
- 경쟁사 ranking 가지는 keyword *우리 미커버* — content gap
- 우리 가지지 않은 schema 종류 — opportunity
- backlink source 차이 — outreach 후보

### Stage 5: ASO 특화 — 평점 / 리뷰 운영

- 평점 4.5+ 유지 — review response 정책
- 부정 review 의 *fix → re-engage* 흐름
- 신규 release 시 *prompt timing* (긍정 경험 직후)

### Stage 6: Drift monitoring

분기 별:
- ranking 변동 (top 10 keyword)
- impression / click rate (Search Console / Sensor Tower)
- conversion rate (visit → action)
- 경쟁사 새 변화

## 5. 산출물 형식

```markdown
## SEO/ASO Audit — {분기}

### Form factor
- web (SEO): {pass / gap}
- mobile (ASO): {pass / gap}

### Technical / On-page (SEO)
| 영역 | pass / gap |

### Metadata (ASO)
| 영역 | pass / gap |

### Keyword research
| keyword | volume | difficulty | intent | priority |

### Content gap
| 경쟁사 keyword | 우리 status |

### Drift trend
- ranking change (분기): ...
- conversion: ...

### Action plan
1. ...
```

## 6. 검증

- [ ] form factor 별 적용 (web SEO / mobile ASO / 둘 다)?
- [ ] SEO 5 영역 (technical / on-page / content / off-page / schema) ?
- [ ] ASO 6 영역 (metadata / visual / description / rating / category / localization) ?
- [ ] Keyword research 4 차원 + intent 분류?
- [ ] Content gap 경쟁사 비교?
- [ ] Drift monitoring 분기 review?

## 7. 다음 phase

- `draft-marketing-copy` — keyword 정합 카피 작성
- `plan-marketing-channel` — paid 보완
- `automate-marketing-content` — content publishing schedule

## 8. 참조

- Google Search Quality Rater Guidelines — E-E-A-T
- App Store Review Guidelines / Google Play Policies
- marketingskills/ai-seo + aso-audit (외부 reference, MIT)
