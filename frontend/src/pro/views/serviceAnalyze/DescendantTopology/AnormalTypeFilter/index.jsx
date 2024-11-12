import { Checkbox, Col, Row, Flex } from 'antd'
import React from 'react'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'

const AnormalTypeFilter = () => {
  const { anormalTypes, setAnormalTypes, anormalTypesOptions } = useAlertAnalyzeContext()
  return (
    <Flex vertical>
      <p
        style={{
          marginBottom: '45px',
          fontSize: '15px',
        }}
      >
        告警和异常总数
      </p>
      <Checkbox.Group
        style={{ width: '100%' }}
        value={anormalTypes}
        onChange={(value) => setAnormalTypes(value)}
      >
        <Row>
          {anormalTypesOptions?.map(
            (option) =>
              !option.hide && (
                <Col key={option.key} span={24}>
                  <Checkbox value={option.value}>{option.label}</Checkbox>
                </Col>
              ),
          )}
        </Row>
      </Checkbox.Group>
    </Flex>
  )
}
export default AnormalTypeFilter
