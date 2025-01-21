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
  return (
    <>
      <Typography>
        <Title level={4}>Zabbix告警接入介绍</Title>
        <Paragraph>
          通过Zabbix的webhook告警媒介,发送告警事件到APO平台. 下面的配置方式适用于Zabbix 7.x版本.
        </Paragraph>
        <Title level={5}>1. 新建告警媒介</Title>
        <Paragraph>
          <ol>
            <li>
              <div>下载媒介配置文件或将下面的配置保存成文件</div>
              <CopyPre code={zabbixExport} />
            </li>
            <li>登录 Zabbix 控制台，选择 `告警(Alert)` {`>`} `媒介(Media Types)`</li>
            <li>
              <div>点击右上角 `导入(Import)` 按钮，选择文件, 选择下载或保存的文件,点击导入</div>
              <Image src={img1}></Image>
            </li>
            <li>
              <div>点击导入好的媒介对象, 修改参数中的 `webhookURL` 为 告警推送地址</div>
              <Image src={img2}></Image>
            </li>
          </ol>
        </Paragraph>

        <Title level={5}>2. 关联告警媒介到用户</Title>
        <Paragraph>推荐使用Admin用户执行告警发送,避免用户权限不足,无法读取到告警事件</Paragraph>
        <Paragraph>
          <ol className="list-decimal">
            <li>在 Zabbix 控制台中, 选择 `用户(User) {`>`} 用户(Users)`</li>
            <li>点击Admin用户,左上角选择 `报警媒介(Media Types)` ,点击 `添加(Add)`</li>
            <li>
              {' '}
              `类型(Type)` 选择 `APO-Collector`, `收件人(Send To)` 填写 `APO` , 点击 `添加(Add)`
              <Image src={img4}></Image>
            </li>
            <li>点击 `更新(Update)`</li>
          </ol>
        </Paragraph>

        <Title level={5}>3. 创建告警动作</Title>
        <Paragraph>推荐使用Admin用户执行告警发送,避免用户权限不足,无法读取到告警事件</Paragraph>

        <Paragraph>
          <ol className="list-decimal">
            <li>在 Zabbix 控制台中, 选择 `告警(Alerts) {`>`} 动作(Actions)`</li>
            <li>右上角点击`创建动作(Create action)`</li>
            <li>`名称(Name)` 填写 `Send To APO`</li>
            <li>选择 `操作(Operations)`, 点击 `操作步骤(Operations)` 中的`添加(Add)`</li>
            <li>
              点击 `发送给用户(Send to users)` 中的 `添加(Add)` , 选中 `Admin` , 再点击 `添加(Add)`
              <Image src={img5}></Image>
            </li>
            <li>
              依次在 `恢复操作(Recovery operations)`, `更新操作(Update operations)` 中重复上述步骤,
              完成后点击添加
            </li>
          </ol>
        </Paragraph>

        <Title level={5}>4. 完成</Title>
        <Paragraph>
          后续可以查询仪表盘中新增问题的动作状态,是否发送成功; 如果动作状态显示`已送达`,
          即可完成发送
        </Paragraph>
        <Image src={img6}></Image>
      </Typography>
    </>
  )
}
export default ZabbixInfo
