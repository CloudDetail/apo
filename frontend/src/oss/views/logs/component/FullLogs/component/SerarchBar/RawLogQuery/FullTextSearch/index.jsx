/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Col, Form, Input, Popover, Row, Space } from 'antd'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { useTranslation } from 'react-i18next' // 引入i18n

const FullTextSearch = () => {
  const { t } = useTranslation('oss/fullLogs')
  const [form] = Form.useForm()
  const { searchValue, setSearchValue, updateQuery } = useLogsContext()
  const [inputValue, setInputValue] = useState()

  const clickSubmit = () => {
    let newQuery = searchValue
    if (newQuery.length > 0) {
      newQuery += ' AND '
    }
    newQuery += '`content` LIKE ' + `'%` + inputValue + `%'`
    setSearchValue(newQuery)
    updateQuery(newQuery)
  }
  return (
    <Space>
      <Input
        placeholder={t('indexList.fullTextSearch.fullTextSearchPlaceholderText')}
        value={inputValue}
        onChange={(e) => setInputValue(e.target.value)}
      />
      <Button type="primary" size="small" onClick={clickSubmit}>
        {t('indexList.fullTextSearch.fullTextSearchText')}
      </Button>
    </Space>
    // <div>
    //   <Form layout="vertical" className="px-2" form={form} initialValues={{ key: 'content' }}>
    //     <Row gutter={10}>
    //       <Col span={7}>
    //         <Form.Item
    //           label="全文检索字段"
    //           name="key"
    //           rules={[
    //             {
    //               required: true,
    //               message: '请输入全文检索字段',
    //             },
    //           ]}
    //         >
    //           <Input placeholder="请输入全文检索字段" />
    //         </Form.Item>
    //       </Col>
    //       <Col span={13}>
    //         <Form.Item
    //           label="全文检索内容"
    //           name="value"
    //           rules={[
    //             {
    //               required: true,
    //               message: '请输入全文检索内容',
    //             },
    //           ]}
    //         >
    //           <Input placeholder="请输入全文检索内容" />
    //         </Form.Item>
    //       </Col>
    //       <Col>
    //         <Form.Item label="  ">
    //           <Button type="primary" htmlType="submit" onClick={clickSubmit}>
    //             确认
    //           </Button>
    //         </Form.Item>
    //       </Col>
    //     </Row>
    //   </Form>
    //   {/* <Button type="primary">全文检索</Button> */}
    // </div>
  )
}
export default FullTextSearch
