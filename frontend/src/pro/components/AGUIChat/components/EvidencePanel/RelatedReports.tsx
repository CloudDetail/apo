import React from 'react'
import { Button, theme } from 'antd'
import { useTranslation } from 'react-i18next'
import { ExternalLink, FileSearch } from 'lucide-react'
import type { FindingsReportLink } from '../../../../api/conversation'

const { useToken } = theme

interface RelatedReportsProps {
  reportLinks?: FindingsReportLink[]
}

interface RelatedReportGroup {
  key: string
  title?: string
  displayCount: number
  links: FindingsReportLink[]
}

const ERROR_REPORT_TITLE_SUFFIX = /\s+error report$/i

const getErrorReportServiceLabel = (link: FindingsReportLink) => {
  if (link.serviceName) {
    return link.serviceName
  }

  if (link.title) {
    const inferredServiceName = link.title.replace(ERROR_REPORT_TITLE_SUFFIX, '').trim()
    return inferredServiceName || link.title
  }

  return undefined
}

const getRelatedReportGroupMeta = (link: FindingsReportLink) => {
  if (link.type !== 'error') {
    return {
      key: `${link.type}:${link.url}`,
      title: undefined,
    }
  }

  const serviceLabel = getErrorReportServiceLabel(link)
  const groupTitle = [serviceLabel, link.errorGroupKey].filter(Boolean).join(' / ')

  return {
    key: `error:${serviceLabel ?? 'unknown'}:${link.errorGroupKey ?? ''}`,
    title: groupTitle || link.title,
  }
}

export const buildRelatedReportGroups = (reportLinks: FindingsReportLink[]) => {
  const groups: RelatedReportGroup[] = []
  const groupByKey = new Map<string, RelatedReportGroup>()

  reportLinks.forEach((link) => {
    const groupMeta = getRelatedReportGroupMeta(link)
    const existingGroup = groupByKey.get(groupMeta.key)

    if (existingGroup) {
      existingGroup.links.push(link)
      existingGroup.displayCount = Math.max(existingGroup.displayCount, link.groupCount ?? 0)
      return
    }

    const group = {
      key: groupMeta.key,
      title: groupMeta.title,
      displayCount: link.groupCount ?? 1,
      links: [link],
    }
    groups.push(group)
    groupByKey.set(group.key, group)
  })

  groups.forEach((group) => {
    group.displayCount = Math.max(group.displayCount, group.links.length)
    group.links.sort((left, right) => {
      if (left.representative === right.representative) {
        return 0
      }

      return left.representative ? -1 : 1
    })
  })

  return groups
}

export const getRelatedReportLabel = ({
  link,
  typeTotal,
  typeIndex,
  groupTotal,
  groupIndex,
  t,
}: {
  link: FindingsReportLink
  typeTotal: number
  typeIndex: number
  groupTotal?: number
  groupIndex?: number
  t: (key: string) => string
}) => {
  if (link.title) {
    const label =
      groupTotal && groupTotal > 1 ? `${link.title} ${groupIndex ?? typeIndex}` : link.title
    return label
  }

  const label =
    link.type === 'error' ? t('evidencePanel.viewErrorReport') : t('evidencePanel.viewSlowReport')

  return typeTotal > 1 ? `${label} ${typeIndex}` : label
}

const RelatedReports: React.FC<RelatedReportsProps> = ({ reportLinks }) => {
  const { token } = useToken()
  const { t } = useTranslation('pro/agui')

  if (!reportLinks?.length) {
    return null
  }

  const typeTotals = reportLinks.reduce<Record<FindingsReportLink['type'], number>>(
    (acc, link) => {
      acc[link.type] += 1
      return acc
    },
    { slow: 0, error: 0 },
  )
  const typeIndexes: Record<FindingsReportLink['type'], number> = {
    slow: 0,
    error: 0,
  }
  const reportGroups = buildRelatedReportGroups(reportLinks)

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <FileSearch size={14} color={token.colorTextSecondary} />
        <span className="text-xs font-medium text-[var(--ant-color-text-secondary)]">
          {t('evidencePanel.relatedReports')}
        </span>
      </div>

      <div className="space-y-2">
        {reportGroups.map((group) => (
          <div key={group.key} className="space-y-1">
            {group.title && (
              <div className="text-[11px] text-[var(--ant-color-text-tertiary)]">
                {group.title}
                {group.displayCount > 1 ? ` · ${group.displayCount}` : ''}
              </div>
            )}
            <div className="flex flex-wrap gap-2">
              {group.links.map((link, groupIndex) => {
                typeIndexes[link.type] += 1
                const label = getRelatedReportLabel({
                  link,
                  typeTotal: typeTotals[link.type],
                  typeIndex: typeIndexes[link.type],
                  groupTotal: group.links.length,
                  groupIndex: groupIndex + 1,
                  t,
                })

                return (
                  <Button
                    key={`${link.type}:${link.url}`}
                    type="link"
                    size="small"
                    href={link.url}
                    target="_blank"
                    rel="noreferrer"
                    icon={<ExternalLink size={14} />}
                    style={{
                      paddingInline: 0,
                    }}
                  >
                    {label}
                  </Button>
                )
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default RelatedReports
