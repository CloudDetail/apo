/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Flex, Form, Tag, Tooltip } from 'antd'
import { ImArrowRight } from 'react-icons/im'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'
import styles from './preview.module.scss'
import classNames from 'classnames'
import { TargetTag } from 'src/core/views/IntegrationCenter/types'
import { Trans, useTranslation } from 'react-i18next'
interface TagRulePreviewProps {
  index: number
}

interface CustomTagProps {
  bordered?: boolean
  color: string
  className?: string
  children: any
}
const CustomTag = ({ bordered = true, color, className = '', children }: CustomTagProps) => {
  return (
    <Tooltip title={children}>
      <Tag
        bordered={bordered}
        color={color}
        className={classNames(styles.ellipsisText, className)} // 动态拼接类名
      >
        {children}
      </Tag>
    </Tooltip>
  )
}

const TagRulePreview = ({ index }: TagRulePreviewProps) => {
  const { t } = useTranslation('core/alertsIntegration')

  const form = Form.useFormInstance()
  const ruleInfo = Form.useWatch(['enrichRuleConfigs', index], form)
  const targetTags = useAlertIntegrationContext((ctx) => ctx.targetTags)
  const getTagNameById = (targetTagId: string) => {
    return targetTags.find((target: TargetTag) => target.id === targetTagId)?.tagName
  }
  console.log(ruleInfo)
  return (
    <>
      {ruleInfo && (
        <>
          {ruleInfo.conditions?.length > 0 && (
            <div>
              {t('tagRulePreview.conditions')}
              {ruleInfo.conditions?.map((condition, index) => (
                <>
                  {index > 0 && <span className="text-[#89ddff] mx-1"> '&&'</span>}
                  <span className="mx-1  text-[#eeffff]">{condition.fromField}</span>
                  <span className="text-[#89ddff]">
                    {condition.operation === 'match' ? '==' : '!=='}{' '}
                  </span>
                  <span className="text-[#c3e88d] mx-1">“{condition.expr}”</span>
                </>
              ))}
            </div>
          )}

          {ruleInfo.rType === 'tagMapping' ? (
            <div className="inline-flex items-center">
              {t('tagRulePreview.tagMapping')}
              <CustomTag bordered={false} color="processing">
                {ruleInfo.fromField}
              </CustomTag>
              ，
              {ruleInfo.fromRegex && (
                <>
                  {t('tagRulePreview.useExpr')}
                  <CustomTag bordered={false} color="cyan">
                    {ruleInfo.fromRegex}
                  </CustomTag>
                  {t('tagRulePreview.mapTo')}
                </>
              )}
              {t('target')}
              <CustomTag bordered={false} color="success">
                {ruleInfo.targetTag.customTag || getTagNameById(ruleInfo.targetTag.targetTagId)}
              </CustomTag>
            </div>
          ) : (
            <>
              <div className="inline-flex items-center">
                <Trans
                  t={t}
                  i18nKey="tagRulePreview.staticEnrichDes"
                  values={{
                    fromField: ruleInfo.fromField,
                    schemaTable: ruleInfo.schemaObject[0],
                    schemaField: ruleInfo.schemaObject[1],
                  }}
                  components={{
                    1: <CustomTag bordered={false} color="processing" />,
                    2: <CustomTag bordered={false} color="geekblue" />,
                  }}
                />
              </div>

              <div className="flex p-2 m-2 border rounded-xl max-w-[600px] justify-center">
                <div className="w-[220px]">
                  <div className="m-2 text-base w-[120px] text-center">{t('extractedField')}</div>
                  <div className="flex items-center justify-center h-[40px]">
                    <CustomTag
                      bordered={false}
                      color="processing"
                      className={`${styles.ellipsisText} w-[120px]  text-sm text-center p-1 `}
                    >
                      {ruleInfo.fromField}
                    </CustomTag>
                    <ImArrowRight className="flex-1" size={30} color="#3f70ff" />
                  </div>
                </div>
                <div className="w-[220px]">
                  <div className={`${styles.ellipsisText} m-2 text-base text-center`}>
                    <Tooltip title={ruleInfo.schemaObject[0]}>{ruleInfo.schemaObject[0]}</Tooltip>
                  </div>
                  <div className="p-2  text-center h-[40px]  border">
                    <Tooltip title={ruleInfo.schemaObject[1]}>{ruleInfo.schemaObject[1]}</Tooltip>
                  </div>
                  {ruleInfo.schemaTargets.map((item) => (
                    <Tooltip title={item.schemaField}>
                      <div
                        className={`${styles.ellipsisText} p-2 text-center h-[40px] border`}
                        //   style={{ border: '.5px solid rgba(150, 219, 12, .5)' }}
                      >
                        {item.schemaField}
                      </div>
                    </Tooltip>
                  ))}
                </div>
                <div className="">
                  <div className="m-2 text-base text-center pl-[80px]">{t('target')}</div>
                  <div className="h-[40px]"></div>
                  {ruleInfo.schemaTargets.map((item) => (
                    <Flex align="center" className="h-[40px] ">
                      <ImArrowRight className="w-[80px]" size={30} color="#96db0b" />
                      <div className="flex-1 text-center">
                        <CustomTag
                          bordered={false}
                          color="success"
                          className={`text-sm text-center px-2 py-1 mr-0 w-[120px]`}
                        >
                          {item.targetTag.customTag || getTagNameById(item.targetTag.targetTagId)}
                        </CustomTag>
                      </div>
                    </Flex>
                  ))}
                </div>
              </div>
            </>
          )}
        </>
      )}
    </>
  )
}
export default TagRulePreview
