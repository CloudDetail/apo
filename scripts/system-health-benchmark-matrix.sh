#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:18080}"
REQUESTS="${REQUESTS:-40}"
CONCURRENCY="${CONCURRENCY:-10}"
WINDOWS="${WINDOWS:-15m 1h 24h}"
OUTPUT_DIR="${OUTPUT_DIR:-output/system-health-benchmark}"
LOG_FILE="${LOG_FILE:-/tmp/apo-backend.log}"
AB_BIN="${AB_BIN:-ab}"

mkdir -p "$OUTPUT_DIR"
RUN_ID="$(date '+%Y%m%d-%H%M%S')"
CSV_FILE="$OUTPUT_DIR/system-health-benchmark-$RUN_ID.csv"
MD_FILE="$OUTPUT_DIR/system-health-benchmark-$RUN_ID.md"

if ! command -v "$AB_BIN" >/dev/null 2>&1; then
  echo "ApacheBench not found. Install ab or set AB_BIN=/path/to/ab." >&2
  exit 1
fi

if [ ! -f "$LOG_FILE" ]; then
  echo "Log file not found: $LOG_FILE" >&2
  exit 1
fi

cat > "$CSV_FILE" <<'CSV'
run_id,endpoint,window,requests,concurrency,failed,rps,mean_ms,mean_all_ms,p50_ms,p90_ms,p95_ms,p99_ms,max_ms,http_200,non_200,cache_hit,cache_miss,cache_shared,current_builds,group_ryg_builds,group_extra_metric_builds
CSV

cat > "$MD_FILE" <<MD
# System Health Benchmark Matrix

- Run ID: \`$RUN_ID\`
- Base URL: \`$BASE_URL\`
- Requests: \`$REQUESTS\`
- Concurrency: \`$CONCURRENCY\`
- Windows: \`$WINDOWS\`
- Log file: \`$LOG_FILE\`

| Endpoint | Window | Failed | RPS | p50 | p95 | p99 | Max | HTTP 200 | Non-200 | Cache hit | Cache miss | Cache shared | Current builds |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
MD

urlencode_path() {
  printf '%s' "$1" | sed 's/ /%20/g'
}

endpoint_path() {
  endpoint="$1"
  window="$2"
  case "$endpoint" in
    overview_summary)
      printf '/api/system-health/overview?summaryOnly=true&skipNewAbnormalMetrics=true&relativeTime=%s' "$window"
      ;;
    systems)
      printf '/api/system-health/systems?page=1&pageSize=50&relativeTime=%s' "$window"
      ;;
    pending_issues)
      printf '/api/system-health/pending-issues?page=1&pageSize=20&relativeTime=%s' "$window"
      ;;
    *)
      echo "unknown endpoint: $endpoint" >&2
      exit 1
      ;;
  esac
}

parse_ab_metric() {
  pattern="$1"
  file="$2"
  awk -v pattern="$pattern" '$0 ~ pattern {print $NF; exit}' "$file"
}

parse_percentile() {
  percentile="$1"
  file="$2"
  awk -v percentile="$percentile" '$1 == percentile"%" {print $2; exit}' "$file"
}

count_logs() {
  pattern="$1"
  start_line="$2"
  end_line="$3"
  if [ "$end_line" -le "$start_line" ]; then
    printf '0'
    return
  fi
  sed -n "$((start_line + 1)),${end_line}p" "$LOG_FILE" | grep -F "$pattern" -c || true
}

count_non_200() {
  start_line="$1"
  end_line="$2"
  if [ "$end_line" -le "$start_line" ]; then
    printf '0'
    return
  fi
  sed -n "$((start_line + 1)),${end_line}p" "$LOG_FILE" | grep -F 'query-request' | grep -F -v '"http_code": 200' -c || true
}

for window in $WINDOWS; do
  for endpoint in overview_summary systems pending_issues; do
    path="$(endpoint_path "$endpoint" "$window")"
    url="$BASE_URL$(urlencode_path "$path")"
    ab_output="$OUTPUT_DIR/ab-$RUN_ID-$endpoint-$window.txt"
    start_line="$(wc -l < "$LOG_FILE" | tr -d ' ')"

    echo "Running $endpoint window=$window url=$url"
    "$AB_BIN" -n "$REQUESTS" -c "$CONCURRENCY" "$url" > "$ab_output"

    end_line="$(wc -l < "$LOG_FILE" | tr -d ' ')"
    failed="$(awk '/Failed requests:/ {print $3; exit}' "$ab_output")"
    rps="$(awk '/Requests per second:/ {print $4; exit}' "$ab_output")"
    mean_ms="$(awk '/Time per request:/ && $0 !~ /across all concurrent/ {print $4; exit}' "$ab_output")"
    mean_all_ms="$(awk '/Time per request:/ && $0 ~ /across all concurrent/ {print $4; exit}' "$ab_output")"
    p50="$(parse_percentile 50 "$ab_output")"
    p90="$(parse_percentile 90 "$ab_output")"
    p95="$(parse_percentile 95 "$ab_output")"
    p99="$(parse_percentile 99 "$ab_output")"
    max_ms="$(awk '$1 == "100%" {print $2; exit}' "$ab_output")"

    http_200="$(count_logs '"http_code": 200' "$start_line" "$end_line")"
    non_200="$(count_non_200 "$start_line" "$end_line")"
    cache_hit="$(count_logs 'cacheStatus="hit"' "$start_line" "$end_line")"
    cache_miss="$(count_logs 'cacheStatus="miss"' "$start_line" "$end_line")"
    cache_shared="$(count_logs 'cacheStatus="shared"' "$start_line" "$end_line")"
    current_builds="$(count_logs 'system-health current systems built' "$start_line" "$end_line")"
    group_ryg_builds="$(count_logs 'system-health overview group ryg built' "$start_line" "$end_line")"
    group_extra_metric_builds="$(count_logs 'system-health overview group extra metrics applied' "$start_line" "$end_line")"

    printf '%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n' \
      "$RUN_ID" "$endpoint" "$window" "$REQUESTS" "$CONCURRENCY" "$failed" "$rps" "$mean_ms" "$mean_all_ms" \
      "$p50" "$p90" "$p95" "$p99" "$max_ms" "$http_200" "$non_200" "$cache_hit" "$cache_miss" "$cache_shared" \
      "$current_builds" "$group_ryg_builds" "$group_extra_metric_builds" >> "$CSV_FILE"

    printf '| %s | %s | %s | %s | %sms | %sms | %sms | %sms | %s | %s | %s | %s | %s | %s |\n' \
      "$endpoint" "$window" "$failed" "$rps" "$p50" "$p95" "$p99" "$max_ms" "$http_200" "$non_200" \
      "$cache_hit" "$cache_miss" "$cache_shared" "$current_builds" >> "$MD_FILE"
  done
done

echo "CSV: $CSV_FILE"
echo "Markdown: $MD_FILE"
