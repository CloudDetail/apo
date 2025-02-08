/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Typography, Image } from 'antd'
import Paragraph from 'antd/es/typography/Paragraph'
import Title from 'antd/es/typography/Title'
import CopyButton from 'src/core/components/CopyButton'
import img1 from 'src/core/assets/alertsIntegration/zabbix/image-1.png'
import img2 from 'src/core/assets/alertsIntegration/zabbix/image-2.png'
import img3 from 'src/core/assets/alertsIntegration/zabbix/image-3.png'
import img4 from 'src/core/assets/alertsIntegration/zabbix/image-4.png'
import img5 from 'src/core/assets/alertsIntegration/zabbix/image-5.png'
import img6 from 'src/core/assets/alertsIntegration/zabbix/image-6.png'
import CopyPre from './CopyPre'
import { useTranslation } from 'react-i18next'
const zabbixExport = `zabbix_export:
  version: '7.0'
  media_types:
    - name: APO-collector
      type: WEBHOOK
      parameters:
        - name: alertId
          value: '{HOST.HOST}-{TRIGGER.ID}'
        - name: alertName
          value: '{TRIGGER.NAME}'
        - name: alertObject
          value: '{ITEM.NAME}'
        - name: alertValue
          value: '{ITEM.VALUE}'
        - name: createTime
          value: '{EVENT.CAUSE.DATE} {EVENT.CAUSE.TIME}'
        - name: detail
          value: '{ALERT.MESSAGE}'
        - name: durationTime
          value: '{EVENT.DURATION}'
        - name: eventTag
          value: '{EVENT.TAGSJSON}'
        - name: host
          value: '{HOST.HOST}'
        - name: ip
          value: '{HOST.IP}'
        - name: recoveryTime
          value: '{EVENT.RECOVERY.DATE} {EVENT.RECOVERY.TIME}'
        - name: severity
          value: '{TRIGGER.NSEVERITY}'
        - name: status
          value: '{TRIGGER.VALUE}'
        - name: updateTime
          value: '{EVENT.DATE} {EVENT.TIME}'
        - name: webhookURL
          value: '<告警推送地址>'
      script: |
        try {
            Zabbix.log(4, '[ APO webhook ] Started with params: ' + value);
            var result = {
                    'tags': {
                        'webhook': 'apo'
                    }
                },
                params = JSON.parse(value),
                req = new HttpRequest(),
                fields = {},
                resp;
            if (params.HTTPProxy) {
                req.setProxy(params.HTTPProxy);
            }
            if (params.Authentication) {
                req.addHeader('Authorization: Basic ' + params.Authentication);
            }
            req.addHeader('Content-Type: application/json');
            result.tags.endpoint = params.webhookURL;
        
            fields.group = 'infra';
            fields.alertId = params.alertId;
            fields.name = params.alertName;
            fields.severity = params.severity;
            fields.status = params.status;
            fields.detail = params.detail;
        
            fields.updateTime = params.updateTime;
            fields.duration = params.durationTime;
            if (params.status == '0') {
                fields.endTime = params.recoveryTime;
            }
        
            var tags = {
                node: params.host,
                node_ip: params.ip,
                alert_object: params.alertObject,
                alert_value: params.alertValue,
            }
        
        	var mergedTags = Object.assign({}, params.eventTag, tags);
            fields.tags = tags;
            resp = req.post(params.webhookURL,
                JSON.stringify(fields)
            );
            if (req.getStatus() != 200) {
                throw 'Response code: ' + req.getStatus();
            }
            result.msg = resp;
            result.tags.status = req.getStatus();
            return JSON.stringify(result);
        }
        catch (error) {
            Zabbix.log(4, '[ APO webhook ] alertEvent push failed json : ' + JSON.stringify(fields));
            Zabbix.log(3, '[ APO webhook ] alertEvent push failed : ' + error);
            throw 'Failed with error: ' + error;
        }
      message_templates:
        - event_source: TRIGGERS
          operation_mode: PROBLEM
          subject: 'Problem: {EVENT.NAME}'
          message: |
            Problem started at {EVENT.TIME} on {EVENT.DATE}
            Problem name: {EVENT.NAME}
            Host: {HOST.NAME}
            Severity: {EVENT.SEVERITY}
            Operational data: {EVENT.OPDATA}
            Original problem ID: {EVENT.ID}
            {TRIGGER.URL}
        - event_source: TRIGGERS
          operation_mode: RECOVERY
          subject: 'Resolved in {EVENT.DURATION}: {EVENT.NAME}'
          message: |
            Problem has been resolved at {EVENT.RECOVERY.TIME} on {EVENT.RECOVERY.DATE}
            Problem name: {EVENT.NAME}
            Problem duration: {EVENT.DURATION}
            Host: {HOST.NAME}
            Severity: {EVENT.SEVERITY}
            Original problem ID: {EVENT.ID}
            {TRIGGER.URL}
`
const ZabbixInfo = () => {
  const { t } = useTranslation('core/alertsIntegration')
  return (
    <>
      <Typography>
        <Title level={4}>{t('zabbixDoc.title')}</Title>
        <Paragraph>{t('zabbixDoc.description')}</Paragraph>

        <Title level={5}>{t('zabbixDoc.step1.title')}</Title>
        <Paragraph>
          <ol>
            <li>
              <div>{t('zabbixDoc.step1.download')}</div>
              <CopyPre code={zabbixExport} />
            </li>
            <li>{t('zabbixDoc.step1.login')}</li>
            <li>
              <div>{t('zabbixDoc.step1.import')}</div>
              <Image src={img1} />
            </li>
            <li>
              <div>{t('zabbixDoc.step1.modifyWebhook')}</div>
              <Image src={img2} />
            </li>
          </ol>
        </Paragraph>

        <Title level={5}>{t('zabbixDoc.step2.title')}</Title>
        <Paragraph>{t('zabbixDoc.step2.recommendation')}</Paragraph>
        <Paragraph>
          <ol className="list-decimal">
            <li>{t('zabbixDoc.step2.navigate')}</li>
            <li>{t('zabbixDoc.step2.selectAdmin')}</li>
            <li>
              {t('zabbixDoc.step2.setType')}
              <Image src={img4} />
            </li>
            <li>{t('zabbixDoc.step2.update')}</li>
          </ol>
        </Paragraph>

        <Title level={5}>{t('zabbixDoc.step3.title')}</Title>
        <Paragraph>{t('zabbixDoc.step3.recommendation')}</Paragraph>
        <Paragraph>
          <ol className="list-decimal">
            <li>{t('zabbixDoc.step3.navigate')}</li>
            <li>{t('zabbixDoc.step3.createAction')}</li>
            <li>{t('zabbixDoc.step3.name')}</li>
            <li>{t('zabbixDoc.step3.selectOperations')}</li>
            <li>
              <div>{t('zabbixDoc.step3.addUser')}</div>
              <Image src={img5} />
            </li>
            <li>{t('zabbixDoc.step3.repeatSteps')}</li>
          </ol>
        </Paragraph>

        <Title level={5}>{t('zabbixDoc.step4.title')}</Title>
        <Paragraph>{t('zabbixDoc.step4.description')}</Paragraph>
        <Image src={img6} />
      </Typography>
    </>
  )
}
export default ZabbixInfo
