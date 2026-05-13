import React from 'react'
import { Button, theme } from 'antd'
import { useTranslation } from 'react-i18next'
import { ExternalLink, FileSearch } from 'lucide-react'
import type { FindingsReportLink } from '../../../../api/conversation'

const { useToken } = theme

interface RelatedReportsProps {
  reportLinks?: FindingsReportLink[]
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

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <FileSearch size={14} color={token.colorTextSecondary} />
        <span className="text-xs font-medium text-[var(--ant-color-text-secondary)]">
          {t('evidencePanel.relatedReports')}
        </span>
      </div>

      <div className="flex flex-wrap gap-2">
        {reportLinks.map((link) => {
          typeIndexes[link.type] += 1
          const label =
            link.type === 'error'
              ? t('evidencePanel.viewErrorReport')
              : t('evidencePanel.viewSlowReport')
          const indexedLabel =
            typeTotals[link.type] > 1 ? `${label} ${typeIndexes[link.type]}` : label

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
              {indexedLabel}
            </Button>
          )
        })}
      </div>
    </div>
  )
}

export default RelatedReports
