import { Checkbox, Col, Row, Flex } from 'antd'
import React from 'react'
import { useServiceAnalyzeContext } from 'src/pro/contexts/ServiceAnalyzeContext'

const AnormalTypeFilter = () => {
  const { anormalTypes, setAnormalTypes, anormalTypesOptions } = useServiceAnalyzeContext()
  return (
    <Flex vertical>
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
