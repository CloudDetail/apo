interface FindingsTextSectionVisibilityInput {
  summaryConclusion?: string
  content?: string
}

export const getFindingsTextSectionVisibility = ({
  summaryConclusion,
  content,
}: FindingsTextSectionVisibilityInput) => {
  const hasSummaryConclusion = Boolean(summaryConclusion)

  return {
    summaryConclusion: hasSummaryConclusion,
    content: !hasSummaryConclusion && Boolean(content),
  }
}
