import { Card, Empty } from 'antd'
import React, { useEffect, useState } from 'react'
import Exceptions from './Exceptions'
import { formatUnit } from 'src/core/utils/time'
import LoadingSpinner from 'src/core/components/Spinner'

const ExceptionTab = ({ node, loading }) => {
  const [errorSpans, setErrorSpans] = useState([])
  useEffect(() => {
    let errorSpans = []
    node.errorSpans?.forEach((errorSpan) => {
      if (errorSpan.exceptions?.length > 0) errorSpans.push(errorSpan)
    })
    setErrorSpans(errorSpans)
  }, [node])
  return errorSpans?.length > 0 ? (
    <React.Fragment>
      {errorSpans.map((item: any, index: number) => (
        <div key={node.id + '-' + item.name + '-' + index}>
          <Card
            className="overflow-hidden"
            title={
              <div className="flex w-full">
                <div>Span： </div>
                <div
                  style={{
                    flex: 1,
                    whiteSpace: 'normal',
                    wordWrap: 'break-word',
                    width: 0,
                  }}
                >
                  {item.name}
                </div>
              </div>
            }
          >
            <div className="error-spans-card">
              <div
                className="flex align-items-center"
                // style={{ color: theme.colors.text.primary }}
              >
                <span style={{ marginRight: 5 }}>Time：</span>
                <span>{formatUnit(item.startTime / 1000000, 'date')}</span>
              </div>
              <div className="flex">
                <span
                  style={{
                    // color: theme.colors.text.primary,
                    margin: '8px 5px 0 0',
                    alignItems: 'baseline',
                  }}
                >
                  Exceptions：
                </span>
                <div style={{ flex: 1, width: 0 }}>
                  <Exceptions exceptions={item.exceptions} spanName={item.name} />
                </div>
              </div>
            </div>
          </Card>
        </div>
      ))}
    </React.Fragment>
  ) : loading ? (
    <LoadingSpinner loading={loading} />
  ) : (
    <Empty />
  )
}
export default ExceptionTab
