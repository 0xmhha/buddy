---
name: design-embedding-search
description: "BM25 + vector embedding + metadata filter + reranking 결합한 hybrid search 설계. write-time embedding + read-time deterministic scoring으로 LLM 호출 최소화. 운영비 절감 + 정확도 양립. feature-management-saas-mcp의 feature.query 같은 semantic 검색 핵심. 트리거: 'embedding 검색 설계' / 'vector search 만들자' / 'BM25 + vector hybrid' / 'semantic 검색' / 'feature.query 구현' / '검색 정확도' / 'rerank 적용'. 입력: 검색 대상 entity + 검색 의도 + 운영비 예산 + 정확도 목표. 출력: index pipeline + query path + rerank 정책 + cost model + cache 전략. 흐름: design-mcp-server → design-embedding-search → build-with-tdd."
type: skill
---

# Design Embedding Search — Hybrid 검색 + 운영비 절감

## 1. 목적

semantic similarity가 필요한 검색 시스템 (feature registry, doc search, support chatbot 등)을 **BM25 + vector + metadata filter + rerank** hybrid로 설계한다.

핵심 가치 2개:
1. **정확도** — keyword만 (BM25) 또는 vector만 (embedding)으로는 부족. hybrid가 표준.
2. **운영비** — 매 query마다 LLM 호출하면 cost 폭발. write-time embedding + read-time deterministic scoring + LLM rerank top-N만.

이 스킬은 `feature-management-saas-mcp.md` "Embedding Knowledge DB와 운영비 절감 정책" 섹션을 직접 구현 가능한 수준으로 변환한다.

## 2. 사용 시점 (When to invoke)

- `feature-management-saas-mcp` `feature.query` MCP tool 구현 전
- product / docs / support article semantic 검색 추가
- Q&A chatbot의 retrieval (RAG) 단계 설계
- 기존 keyword 검색을 hybrid로 upgrade
- 운영비 폭증 시 (LLM 호출 줄이기)
- multi-language 검색 (영어 + 한국어 + 일본어)
- 신규 entity type 추가 시 index pipeline 확장

## 3. 입력 (Inputs)

### 필수
- 검색 대상 entity (feature / doc / 사용자 / product)
- 검색 의도 (find similar / find exact / find candidates)
- 운영비 예산 ($/월, query당 cost 목표)
- 정확도 목표 (recall@5 / NDCG / MRR)

### 선택
- 기존 검색 시스템 (Elasticsearch, OpenSearch)
- 사용 LLM (Claude / GPT / open-source)
- 데이터 규모 (entity 수 / query QPS)

### 입력 부족 시 forcing question
- "query당 비용 목표가 얼마야? $0.001? $0.01? rerank 깊이 결정 영향."
- "정확도 평가 dataset 있어? labeled relevant 결과 없으면 튜닝 못함."
- "multi-language 필요해? 다국어 embedding model 선택 다름."
- "data 규모 얼마야? 10K 미만이면 brute-force OK. 100K+ 면 ANN index."

## 4. 핵심 원칙 (Principles)

1. **Write-time embedding** — 검색 시점이 아닌 등록 / 변경 시점에 생성. 매 query 호출 금지.
2. **Hybrid first** — BM25 + vector 병행. 어느 한쪽 단독은 약함.
3. **Metadata filter는 hard gate** — license / visibility / language 같은 binary는 전 단계 적용.
4. **Reranking은 top-N만** — top 50 → top 5 reranking. 전체에 LLM 적용 금지.
5. **Deterministic scoring 우선** — semantic + interface + stack은 vector / structural 비교. LLM은 explanation에만.
6. **Cache** — query normalization + result cache. 중복 query 비용 0.
7. **Changed field만 reindex** — feature 1개 변경 시 그 feature embedding만 갱신. 전체 reindex 금지.
8. **Cold path vs hot path** — 자주 조회되는 feature는 advanced index. tail은 BM25만으로 충분.

## 5. 단계 (Phases)

### Phase 1. Schema Design
검색 대상 entity의 검색 가능 field:
- text fields (name, summary, problem, description)
- structured fields (stack, language, license, status)
- vector fields (embedding from text)
- metadata (visibility, owner, version)

### Phase 2. Write Path (Indexing)
```
entity store/update
  → source hash check (변경 감지)
  → deterministic parser (license, language, stack 추출)
  → lightweight profile generation
  → changed field embedding 생성 (LLM 또는 SaaS embedding API)
  → vector DB upsert
  → BM25 index update
  → metadata filter index update
```

### Phase 3. Read Path (Query)
```
query
  → query normalization (lowercase, stemming, alias)
  → cache lookup
  → metadata filter (hard gate)
  → BM25 search (top 100)
  → vector search (top 100)
  → fusion (RRF / weighted)
  → top 50 후보
  → optional LLM rerank top 50 → top 5
  → return + explanation
```

### Phase 4. Embedding Model Selection
- text-embedding-3-large (OpenAI) — 1536/3072 dim
- voyage-3 (Voyage AI) — 영어 강함
- multilingual-e5 (open-source) — 다국어
- BGE / E5 — 자체 호스팅

비용 vs 품질 trade-off.

### Phase 5. Vector Index
- pgvector (PostgreSQL extension) — 작은 규모, transactional
- Pinecone / Weaviate / Qdrant — managed, 큰 규모
- Milvus — 자체 호스팅, GPU
- HNSW vs IVF — recall vs latency trade-off

### Phase 6. Fusion & Reranking
- Reciprocal Rank Fusion (RRF) — 단순, 효과적
- weighted score (BM25 * w1 + vector * w2)
- LLM rerank — top 50 → top 5, prompt에 query + 후보 N개

### Phase 7. Caching & Cost Optimization
- query normalization → cache key
- result TTL (5분 / 1시간 / 1일 — entity 변동성 따라)
- write-time embedding cache (hash 기반 dedup)
- changed field만 reindex (cost 80% 절감)
- cold path는 LLM rerank skip

### Phase 8. Evaluation
- offline: labeled dataset, recall@5, NDCG
- online: A/B test, click-through, user feedback
- regression test (정확도 + latency)

## 6. 출력 템플릿 (Output Format)

```yaml
search_system:
  target_entity: feature
  expected_qps: 100
  expected_volume: 1M features
  budget_per_query: "$0.005"
  accuracy_target:
    recall_5: 0.85
    ndcg_10: 0.75
    p95_latency_ms: 500

schema:
  text_fields: [name, summary, problem, scope, primary_flow]
  structured_fields: [stack, language, license, status, visibility]
  vector_fields:
    - name: feature_profile_embedding
      source: concat(name, summary, problem)
      model: text-embedding-3-large
      dimension: 1536
  metadata_filters: [license, visibility, language, status]

write_path:
  trigger: feature.store / feature.update
  steps:
    - source_hash_check
    - deterministic_parser
    - profile_generation (template-based, no LLM)
    - changed_field_embedding (only if text_changed)
    - vector_upsert
    - bm25_index_update
  cost_per_write: "$0.0002 (avg)"
  reindex_full: weekly  # bg job
  reindex_changed: real_time

read_path:
  steps:
    - query_normalization
    - cache_lookup (hit ratio target 30%+)
    - metadata_filter
    - bm25_search (top 100)
    - vector_search (top 100)
    - fusion (RRF, k=60)
    - candidates (top 50)
    - llm_rerank (top 50 → top 5, optional)
    - explanation_generation (top 5, LLM)
  cost_per_query:
    cached: "$0.00001"
    bm25_vector_only: "$0.0005"
    with_rerank: "$0.005"
    with_explanation: "$0.01"

embedding_model:
  primary: text-embedding-3-large
  dimension: 1536
  multilingual: yes
  pricing: "$0.13 per 1M tokens"
  alternatives:
    - voyage-3 (영어 강함)
    - multilingual-e5 (자체 호스팅)

vector_index:
  type: pgvector  # or Pinecone | Qdrant | Weaviate
  index_method: HNSW
  params:
    m: 16
    ef_construction: 64
    ef_search: 40
  recall_target: 0.95

fusion_strategy:
  method: reciprocal_rank_fusion
  k: 60
  weight_bm25: 1.0
  weight_vector: 1.0

rerank_policy:
  enabled_by_default: yes
  threshold: top_50
  output: top_5
  llm: claude-haiku-4-5  # cheaper
  prompt_template: "<feature_reuse_review>"
  bypass_for:
    - tail_features (popularity < threshold)
    - exact_match_high_confidence

caching:
  query_cache:
    backend: redis
    ttl_seconds: 300
    keying: normalized_query + filters_hash
  embedding_cache:
    backend: postgres
    ttl_seconds: infinite  # only invalidate on source change
    keying: source_hash
  hit_ratio_target: 30

cost_model:
  monthly_estimate:
    queries: 1M
    cached: 30%
    bm25_vector_only: 60%
    with_rerank: 10%
    total: "$3500/month"
  optimization_levers:
    - cache_ttl_increase (hit ratio +10% = $300 saved)
    - rerank_threshold_raise (rerank ratio -5% = $250 saved)
    - tail_skip_rerank ($150 saved)

evaluation:
  offline:
    dataset: labeled_relevance.jsonl
    metrics: [recall@5, ndcg@10, mrr]
    frequency: weekly
  online:
    metric: ctr_top_5
    a_b_test_required: yes
    regression_alert: ndcg drop >5%

unknown_term_handling:
  query_expansion:
    method: llm_synonyms
    cache: yes
  unknown_term_log: yes  # for periodic dictionary update
  inferred_meaning: returned_with_confidence

multilingual:
  supported: [ko, en, ja]
  embedding_model: multilingual-e5 OR text-embedding-3-large
  bm25_analyzer: per_language (kuromoji, mecab, default)
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `design-mcp-server` — `Skill` tool로 invoke (`feature.query` 정의)
- 페어: `design-billing-system` — query 비용 / quota
- 페어: `audit-security` — input validation, prompt injection
- 다음 단계: `build-with-tdd` — relevance test (labeled dataset)
- 다음 단계: `monitor-regressions` — accuracy / latency 회귀 모니터

## 8. Anti-patterns

1. **Query마다 LLM 호출** — 운영비 폭증. write-time embedding + cache.
2. **Vector only / BM25 only** — 한쪽만 약함. hybrid가 표준.
3. **전체 reranking** — top 50만 rerank. 1000개 rerank하면 $$$.
4. **전체 reindex** — feature 1개 변경에 전체 재처리. changed field만.
5. **Cache 없음** — 동일 query 반복. 30% hit ratio 쉬움.
6. **Metadata filter를 score에 포함** — license는 hard gate. score에 섞으면 위반 결과 노출.
7. **Cold path에 advanced index** — tail features 99%는 단순 검색. cost vs 품질 trade-off.
8. **Evaluation dataset 없음** — 정확도 모름. labeled dataset + offline + online 강제.
