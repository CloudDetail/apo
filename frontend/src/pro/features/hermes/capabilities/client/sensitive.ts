const SENSITIVE_NAME =
  '(?:password|passwd|pwd|token|secret|credential|api[\\s_-]*key|access[\\s_-]*key|private[\\s_-]*key|密码|口令|令牌|密钥|秘钥|凭据|私钥|证书)'

const ASSIGNMENT_PATTERN = new RegExp(
  `(${SENSITIVE_NAME}\\s*(?:[:=：]|\\b(?:is|equals|set\\s+to)\\b|(?:是|为|等于|设置为))\\s*)(["'“”‘’\\x60]?[^\\s,，;；。！？\\n]+)`,
  'gi',
)

const ASSIGNMENT_TEST_PATTERN = new RegExp(ASSIGNMENT_PATTERN.source, 'i')
const LONG_TOKEN_PATTERN = new RegExp(
  `\\b${SENSITIVE_NAME}\\s+[A-Za-z0-9][A-Za-z0-9._~+/=-]{11,}\\b`,
  'i',
)
const BEARER_PATTERN = /\b(?:bearer|basic)\s+[A-Za-z0-9._~+/=-]{12,}\b/i

const SENSITIVE_HINT_PATTERN = new RegExp(SENSITIVE_NAME, 'i')
const BENIGN_VALUE_PATTERN = /^(?:expired|invalid|missing|none|null|empty|unknown|过期|失效|无效|缺失|未配置|未提供|空|未知)$/i

/** Redacts values while keeping the field name visible for safe diagnostics. */
export function redactSensitiveText(value: string): string {
  return value
    .replace(ASSIGNMENT_PATTERN, '$1[已隐藏]')
    .replace(LONG_TOKEN_PATTERN, (match) => match.replace(/\s+\S+$/u, ' [已隐藏]'))
    .replace(BEARER_PATTERN, (match) => match.replace(/(\s+).+$/u, '$1[已隐藏]'))
}

/** Returns true only when a sensitive field appears to carry a value. */
export function containsSensitiveValue(value: string): boolean {
  const assignment = value.match(ASSIGNMENT_TEST_PATTERN)
  if (assignment && !BENIGN_VALUE_PATTERN.test(assignment[2].replace(/["'“”‘’`]/g, ''))) return true
  return LONG_TOKEN_PATTERN.test(value) || BEARER_PATTERN.test(value)
}

export function isSensitiveHint(value: string): boolean {
  return SENSITIVE_HINT_PATTERN.test(value)
}
