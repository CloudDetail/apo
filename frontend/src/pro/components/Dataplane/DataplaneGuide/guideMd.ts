import Datadog1 from 'src/pro/assets/dataplane/Datadog-1.png'
import Datadog2 from 'src/pro/assets/dataplane/Datadog-2.png'
import Datadog4 from 'src/pro/assets/dataplane/Datadog-4.png'
import Datadog5 from 'src/pro/assets/dataplane/Datadog-5.png'
import Datadog9 from 'src/pro/assets/dataplane/Datadog-9.png'
import Datadog11 from 'src/pro/assets/dataplane/Datadog-11.png'
import GrafanaTempoCloud1 from 'src/pro/assets/dataplane/GrafanaTempoCloud-1.png'
import GrafanaTempoCloud2 from 'src/pro/assets/dataplane/GrafanaTempoCloud-2.png'
import GrafanaTempoCloud3 from 'src/pro/assets/dataplane/GrafanaTempoCloud-3.png'
import GrafanaTempoCloud4 from 'src/pro/assets/dataplane/GrafanaTempoCloud-4.png'
import GrafanaTempoCloud5 from 'src/pro/assets/dataplane/GrafanaTempoCloud-5.png'
import GrafanaTempoCloud6 from 'src/pro/assets/dataplane/GrafanaTempoCloud-6.png'
import AlicloudArms1 from 'src/pro/assets/dataplane/AlicloudArms-1.png'
import AlicloudArms2 from 'src/pro/assets/dataplane/AlicloudArms-2.png'
import AlicloudArms3 from 'src/pro/assets/dataplane/AlicloudArms-3.png'
import AlicloudArms4 from 'src/pro/assets/dataplane/AlicloudArms-4.png'
import AlicloudArms5 from 'src/pro/assets/dataplane/AlicloudArms-5.png'
import GrafanaLoki1 from 'src/pro/assets/dataplane/GrafanaTempoCloud-1.png'
import GrafanaLoki2 from 'src/pro/assets/dataplane/GrafanaTempoCloud-2.png'

const guideMdMapEn = {
    'datadog': `### Step 1: Access Your Datadog Account

+ Open your Datadog URL and log in to your Datadog account.
+ Confirm the \`Site Parameter\` for your site URL according to the following table (Ref. from [official Datadog website](https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site)), and enter it into the \`Datadog Site\` filed. For example, if your site URL is \`https://app.datadoghq.com\`, then enter \`datadoghq.com\`.

| Site | Site URL | Site Parameter | Location |
| --- | --- | --- | --- |
| US1 | \`https://app.datadoghq.com\` | \`datadoghq.com\` | US |
| US3 | \`https://us3.datadoghq.com\` | \`us3.datadoghq.com\` | US |
| US5 | \`https://us5.datadoghq.com\` | \`us5.datadoghq.com\` | US |
| EU1 | \`https://app.datadoghq.eu\` | \`datadoghq.eu\` | EU (Germany) |
| US1-FED | \`https://app.ddog-gov.com\` | \`ddog-gov.com\` | US |
| AP1 | \`https://ap1.datadoghq.com\` | \`ap1.datadoghq.com\` | Japan |
| AP2 | \`https://ap2.datadoghq.com\` | \`ap2.datadoghq.com\` | Australia |

### Step 2: Create Datadog API Key

1. Hover over the user avatar in the bottom left, and click \`API Keys\` under the \`Organization Settings\` section. ![](${Datadog1})
2. Click \`+ New Key\`, enter a name for the API Key, and confirm. ![](${Datadog2})
3. Copy the newly created API Key and enter it into the \`Datadog API Key\` field.

### Step 3: Create Application Key

1. Hover over the user avatar in the bottom left, and click \`Application Keys\` under the \`Organization Settings\` section. ![](${Datadog4})
2. Click \`+ New Key\`, enter a name for the Application Key, and confirm. ![](${Datadog5})
3. Copy the newly created Application Key and enter it into the \`Datadog APP Key\` field. 

### (Optional) Step 4: Configure Application Keys Permissions

Application Keys default to \`Not Scoped\`, meaning full access is granted, and users can use them without any adjustments. If you want to implement minimal permission management, Syncause data integration requires Application Keys to have the following permissions:

+ \`apm_read\`
+ \`timeseries_query\`
+ \`logs_read_data\`
+ \`logs_read_index_data\`

#### Edit Application Keys Permissions

+ Click the \`Edit\` button next to \`Scope\`.
+ In the scope list, select the minimal permissions as described in the section above, and click save.

![](${Datadog9})

### Step 5: Configure Datadog env  
Identify which applications you want to integrate with Syncause.

1. Click \`APM\` from the left-hand menu, then click \`Services\` on the right-hand card.
2. Locate the environment value for the application to be monitored, and enter that value into the \`Datadog env\` field.
3. If you have multiple environments (for example, dev, staging, prod), create a separate Datadog integration for each.

![](${Datadog11})`,
    'jaeger': `### Step 1: Access Your Jaeger UI

1. Open your Jaeger UI. By default, it is available at \`http://IP:16686\`. If you have customized the QUERY_BASE_PATH, make sure to use the updated path.
2. Copy the Jaeger UI address without the \`/search\` suffix, and enter it into the \`Jaeger Address\` field.

For example, if your Jaeger UI address is \`http://10.0.0.0:16686/search\`, enter \`http://10.0.0.0:16686\`.`,
    'prometheusmetric': `
This integration requires only basic connection parameters. Simply fill in the following fields with information from your environment:

1. **Address**: The service endpoint or API address for your data source (for example, Tempo, Prometheus, or SkyWalking).
2. **Username / Password / Token** (If applicable): The authentication credentials required to access the service.

Once all fields are completed, click **Save** to register the integration.`,
    'tempo': `### Step 1: Access your Grafana Cloud account

1. Access to Grafana Cloud website and login your account.
2. Click \`My Account\` in the upper right corner. ![](${GrafanaTempoCloud1})
3. Click \`Details\` under your Grafana Cloud stack section. ![](${GrafanaTempoCloud2})
4. Click \`Details\` in \`Tempo\` section. ![](${GrafanaTempoCloud3})

### Step 2: Get your tempo information

1. Scroll down to \`Grafana Data Source settings\` section.
2. Generate an \`Password\` if you don't have one.
3. Copy and paste the following values into the corresponding fields:
   - \`URL\` value into \`Tempo Address\` field; 
   - \`User\` value into \`Tempo Username\` field; 
   - \`Password\` value into \`Tempo Token\` field.

![](${GrafanaTempoCloud4})

### (Optional) Step 3: Set up Prometheus integration for topology data

Grafana Cloud Tempo does not automatically generate Service Graph (topology) data. If the topology is missing, you can enable **Observability** -> **Application** to generate it. Please note that this may incur additional charges.

If Service Graph (topology) data is already available and you want to integrate it, please follow these steps:

1. Open your Grafana Cloud stack and click \`Details\` in the \`Prometheus\` section. ![](${GrafanaTempoCloud5})
2. Scroll down to \`Query Endpoint\` section.
3. Generate an \`Password\` if you don't already have one.
4. Copy and paste the following values into the corresponding fields:
   - \`Query URL\` value into \`PrometheusAddress\` field; 
   - \`Username\` value into \`Prometheus Username\` field; 
   - \`Password\` value into \`Prometheus Token\` field.

![](${GrafanaTempoCloud6})`,
    'arms': `### Step 1: Confirm ARMS Service Endpoint

1. Visit the [ARMS Service Endpoint documentation](https://www.alibabacloud.com/help/arms/application-monitoring/developer-reference/api-arms-2019-08-08-endpoint-apps).
2. Locate the \`Public network access address\` corresponding to your region and copy it.
3. Enter the address value into the \`ARMS Address\` field in the configuration panel.

### Step 2: Create a RAM User for Integration

To securely connect Syncause with ARMS, create a dedicated RAM user with minimal permissions.

1. Go to the [Alibaba Cloud RAM User Console](https://ram.console.aliyun.com/overview?activeTab=overview). ![](${AlicloudArms1})
2. Click **Create User**, set a login name, and check the **OpenAPI Access** option. ![](${AlicloudArms2})
3. After creation, copy the **AccessKey ID** and **AccessKey Secret**. ![](${AlicloudArms3})
4. Enter these values into the corresponding fields:
   * **AccessKey ID** → \`ARMS AccessKey\` field
   * **AccessKey Secret** → \`ARMS AccessSecret\` field

### Step 3: Assign Minimal Permissions

1. In the RAM User Console, open the newly created user's details page.
2. Click **Add Permissions**. ![](${AlicloudArms4})
3. Select the \`AliyunARMSReadOnlyAccess\` policy and click **Grant Permissions**.
4. Wait for the operation to complete. ![](${AlicloudArms5})`,
    'signoz': `### Step 1: Locate Your SigNoz API Endpoint

1. Log in to your SigNoz instance.
2. Identify your SigNoz instance URL (for example, \`example.signoz.io\`).
3. Construct your API endpoint in the following format:
   \`\`\`
   https://{URL}/api/v5/query_range
   \`\`\`
   Example:

   \`\`\`
   https://example.signoz.io/api/v5/query_range
   \`\`\`
4. Enter the endpoint value into the \`SigNoz Address\` field in the configuration panel.

### Step 2: Create a SigNoz API Key

1. Log in to your SigNoz instance.
2. Navigate to the \`Settings\` page in the SigNoz UI.
3. Click **+ New Key**, enter a name for the API key, and confirm.
4. Copy the generated API key and paste it into the \`SigNoz API Key\` field in the configuration panel.
   * Only users with the **Admin** role can create or manage API keys.
   * If you do not have admin privileges, please contact your organization's administrator.`,
    'elastic': `### Scope

This page currently exposes **Elastic / ELK Log** onboarding only.

If your team usually calls this stack **ELK**, APO connects to the **Elasticsearch** endpoint directly, not the Kibana URL.

### Step 1: Get the Elasticsearch endpoint

1. Use the Elasticsearch HTTPS endpoint directly.
2. Elastic Cloud users can copy it from **Deployments** -> **Elasticsearch**.
3. Self-managed users can use an address such as \`https://elastic.example.com:9200\`.

### Step 2: Create an API Key

1. Open **Stack Management** -> **API keys** in Elastic Cloud / Kibana, or create an Elasticsearch API key in your self-managed cluster.
2. The key only needs read permission on the target log indices or data streams, for example \`logs-*\`.

### Step 3: Fill APO and test

1. Only fill in **Name**, **URL**, and **API Key** in the main flow. The **URL** here is the Elasticsearch **Address**.
2. APO uses these defaults automatically: \`authType=apiKey\`, \`logIndexPattern=logs-*\`.
3. Click **Test** and then **Connect**.
4. For AgentSRE service log analysis, your Elastic logs should include common Kubernetes fields such as \`kubernetes.pod.name\` / \`k8s.pod.name\` and \`kubernetes.namespace\` / \`k8s.namespace.name\`. Standard Elastic Agent, Filebeat, and OpenTelemetry Kubernetes logs usually provide these fields automatically.
5. A warning means the connection is valid, but APO could not confirm that the current log samples can be filtered by service instance. You can still save the integration, but AgentSRE may not find service-specific business logs until those fields are present.`,
    'skywalking': `### Configuration

This integration requires only basic connection parameters. Simply fill in the following fields with information from your environment:

1. **Address**: The service endpoint or API address for your data source (for example, Tempo, Prometheus, or SkyWalking).
2. **Username / Password / Token** (If applicable): The authentication credentials required to access the service.

Once all fields are completed, click **Save** to register the integration.`,
    'loki': `### Step 1: Access your Grafana Cloud account

1. Access to Grafana Cloud website and login your account.
2. Click \`My Account\` in the upper right corner. ![](${GrafanaLoki1})
3. Click \`Details\` under your Grafana Cloud stack section. ![](${GrafanaLoki2})
4. Click \`Details\` in \`Loki\` section.

### Step 2: Get your Loki instance details

1. Scroll down to \`Grafana Data Source settings\` section.
2. Generate an \`Password\` if you don't already have one.
3. Copy and paste the following values into the corresponding fields:
   - \`URL\` value into \`Loki Address\` field; 
   - \`User\` value into \`Loki Username\` field; 
   - \`Password\` value into \`Loki Token\` field.
`,
    'grafanaPrometheus': `### Step 1: Access your Grafana Cloud account

1. Access to Grafana Cloud website and login your account.
2. Click \`My Account\` in the upper right corner. ![](${GrafanaTempoCloud1})
3. Click \`Details\` under your Grafana Cloud stack section. ![](${GrafanaTempoCloud2})
4. Click \`Details\` in \`Prometheus\` section.

### Step 2: Get your Prometheus instance details

1. Scroll down to \`Query Endpoint\` section.
2. Generate an \`Password\` if you don't already have one.
3. Copy and paste the following values into the corresponding fields:
   - \`Query URL\` value into \`PrometheusAddress\` field; 
   - \`Username\` value into \`Prometheus Username\` field; 
   - \`Password\` value into \`Prometheus Token\` field.

![](${GrafanaTempoCloud6})`,
}

const guideMdMapZh = {
    'datadog': `### 步骤 1：访问您的 Datadog 账户

+ 打开您的 Datadog URL 并登录您的 Datadog 账户。
+ 根据下表确认您站点 URL 的 \`Site Parameter\`（参考自 [Datadog 官方网站](https://docs.datadoghq.com/getting_started/site/#access-the-datadog-site)），并将其输入到 \`Datadog Site\` 字段中。例如，如果您的站点 URL 是 \`https://app.datadoghq.com\`，则输入 \`datadoghq.com\`。

| 站点 | 站点 URL | 站点参数 | 位置 |
| --- | --- | --- | --- |
| US1 | \`https://app.datadoghq.com\` | \`datadoghq.com\` | 美国 |
| US3 | \`https://us3.datadoghq.com\` | \`us3.datadoghq.com\` | 美国 |
| US5 | \`https://us5.datadoghq.com\` | \`us5.datadoghq.com\` | 美国 |
| EU1 | \`https://app.datadoghq.eu\` | \`datadoghq.eu\` | 欧盟（德国） |
| US1-FED | \`https://app.ddog-gov.com\` | \`ddog-gov.com\` | 美国 |
| AP1 | \`https://ap1.datadoghq.com\` | \`ap1.datadoghq.com\` | 日本 |
| AP2 | \`https://ap2.datadoghq.com\` | \`ap2.datadoghq.com\` | 澳大利亚 |

### 步骤 2：创建 Datadog API Key

1. 将鼠标悬停在左下角的用户头像上，然后点击 \`Organization Settings\` 部分下的 \`API Keys\`。 ![](${Datadog1})
2. 点击 \`+ New Key\`，输入 API Key 的名称并确认。 ![](${Datadog2})
3. 复制新创建的 API Key 并将其输入到 \`Datadog API Key\` 字段中。

### 步骤 3：创建 Application Key

1. 将鼠标悬停在左下角的用户头像上，然后点击 \`Organization Settings\` 部分下的 \`Application Keys\`。 ![](${Datadog4})
2. 点击 \`+ New Key\`，输入 Application Key 的名称并确认。 ![](${Datadog5})
3. 复制新创建的 Application Key 并将其输入到 \`Datadog APP Key\` 字段中。

### （可选）步骤 4：配置 Application Keys 权限

Application Keys 默认为 \`Not Scoped\`，意味着授予完全访问权限，用户可以直接使用而无需任何调整。如果您想实施最小权限管理，Syncause 数据集成需要 Application Keys 具有以下权限：

+ \`apm_read\`
+ \`timeseries_query\`
+ \`logs_read_data\`
+ \`logs_read_index_data\`

#### 编辑 Application Keys 权限

+ 点击 \`Scope\` 旁边的 \`Edit\` 按钮。
+ 在权限列表中，选择上述部分描述的最小权限，然后点击保存。

![](${Datadog9})

### 步骤 5：配置 Datadog env  
确定您想要与 Syncause 集成的应用程序。

1. 从左侧菜单点击 \`APM\`，然后点击右侧卡片上的 \`Services\`。
2. 找到要监控的应用程序的环境值，并将该值输入到 \`Datadog env\` 字段中。
3. 如果您有多个环境（例如 dev、staging、prod），请为每个环境创建单独的 Datadog 集成。

![](${Datadog11})`,
    'jaeger': `### 步骤 1：访问您的 Jaeger UI

1. 打开您的 Jaeger UI。默认情况下，它位于 \`http://IP:16686\`。如果您自定义了 QUERY_BASE_PATH，请确保使用更新后的路径。
2. 复制 Jaeger UI 地址（不包含 \`/search\` 后缀），并将其输入到 \`Jaeger Address\` 字段中。

例如, 如果您的 Jaeger UI 地址是 \`http://10.0.0.0:16686/search\`, 则输入 \`http://10.0.0.0:16686\`.`,
    'prometheusmetric': `
此集成仅需要基本的连接参数。只需根据您的环境信息填写以下字段：

1. **地址**：数据源的服务端点或 API 地址（例如 Tempo、Prometheus 或 SkyWalking）。
2. **用户名 / 密码 / Token**（如适用）：访问服务所需的身份验证凭据。

填写完所有字段后，点击 **保存** 以注册集成。`,
    'tempo': `### 步骤 1：访问您的 Grafana Cloud 账户

1. 访问 Grafana Cloud 网站并登录您的账户。
2. 点击右上角的 \`My Account\`。 ![](${GrafanaTempoCloud1})
3. 点击您的 Grafana Cloud stack 部分下的 \`Details\`。 ![](${GrafanaTempoCloud2})
4. 点击 \`Tempo\` 部分的 \`Details\`。 ![](${GrafanaTempoCloud3})

### 步骤 2：获取您的 Tempo 信息

1. 向下滚动到 \`Grafana Data Source settings\` 部分。
2. 如果您还没有 \`Password\`，请生成一个。
3. 复制并粘贴以下值到相应的字段：
   - 将 \`URL\` 值输入到 \`Tempo Address\` 字段；
   - 将 \`User\` 值输入到 \`Tempo Username\` 字段；
   - 将 \`Password\` 值输入到 \`Tempo Token\` 字段。

![](${GrafanaTempoCloud4})

### （可选）步骤 3：设置 Prometheus 集成以获取拓扑数据

Grafana Cloud Tempo 不会自动生成 Service Graph（拓扑）数据。如果缺少拓扑数据，您可以启用 **Observability** -> **Application** 来生成它。请注意，这可能会产生额外费用。

如果 Service Graph（拓扑）数据已经可用并且您想要集成它，请按照以下步骤操作：

1. 打开您的 Grafana Cloud stack，点击 \`Prometheus\` 部分的 \`Details\`。 ![](${GrafanaTempoCloud5})
2. 向下滚动到 \`Query Endpoint\` 部分。
3. 如果您还没有 \`Password\`，请生成一个。
4. 复制并粘贴以下值到相应的字段：
   - 将 \`Query URL\` 值输入到 \`PrometheusAddress\` 字段；
   - 将 \`Username\` 值输入到 \`Prometheus Username\` 字段；
   - 将 \`Password\` 值输入到 \`Prometheus Token\` 字段。

![](${GrafanaTempoCloud6})`,
    'arms': `### 步骤 1：确认 ARMS 服务端点

1. 访问 [ARMS 服务端点文档](https://www.alibabacloud.com/help/arms/application-monitoring/developer-reference/api-arms-2019-08-08-endpoint-apps)。
2. 找到与您的地区对应的 \`公网访问地址\` 并复制它。
3. 将地址值输入到配置面板中的 \`ARMS Address\` 字段。

### 步骤 2：创建用于集成的 RAM 用户

为了安全地将 Syncause 与 ARMS 连接，请创建一个具有最小权限的专用 RAM 用户。

1. 访问 [阿里云 RAM 用户控制台](https://ram.console.aliyun.com/overview?activeTab=overview)。 ![](${AlicloudArms1})
2. 点击 **创建用户**，设置登录名称，并勾选 **OpenAPI 访问** 选项。 ![](${AlicloudArms2})
3. 创建后，复制 **AccessKey ID** 和 **AccessKey Secret**。 ![](${AlicloudArms3})
4. 将这些值输入到相应的字段：
   * **AccessKey ID** → \`ARMS AccessKey\` 字段
   * **AccessKey Secret** → \`ARMS AccessSecret\` 字段

### 步骤 3：分配最小权限

1. 在 RAM 用户控制台中，打开新创建用户的详情页面。
2. 点击 **添加权限**。 ![](${AlicloudArms4})
3. 选择 \`AliyunARMSReadOnlyAccess\` 策略并点击 **授权**。
4. 等待操作完成。 ![](${AlicloudArms5})`,
    'signoz': `### 步骤 1：确定您的 SigNoz API 端点

1. 登录到您的 SigNoz 实例。
2. 确定您的 SigNoz 实例 URL（例如 \`example.signoz.io\`）。
3. 按以下格式构建您的 API 端点：
   \`\`\`
   https://{URL}/api/v5/query_range
   \`\`\`
   示例：

   \`\`\`
   https://example.signoz.io/api/v5/query_range
   \`\`\`
4. 将端点值输入到配置面板中的 \`SigNoz Address\` 字段。

### 步骤 2：创建 SigNoz API Key

1. 登录到您的 SigNoz 实例。
2. 在 SigNoz UI 中导航到 \`Settings\` 页面。
3. 点击 **+ New Key**，输入 API 密钥的名称并确认。
4. 复制生成的 API 密钥并将其粘贴到配置面板中的 \`SigNoz API Key\` 字段。
   * 只有具有 **Admin** 角色的用户才能创建或管理 API 密钥。
   * 如果您没有管理员权限，请联系您组织的管理员。`,
    'elastic': `### 当前范围

当前页面先开放 **Elastic / ELK 日志 Log 接入**。

如果你们内部习惯叫 **ELK**，这里对应的是 **Elasticsearch** 接入，不是 Kibana 页面地址。

### 步骤 1：获取 Elasticsearch 地址

1. 请直接使用 Elasticsearch 的 HTTPS endpoint。
2. Elastic Cloud 用户可在 **Deployments** -> **Elasticsearch** 中复制 endpoint。
3. 自建集群可使用类似 \`https://elastic.example.com:9200\` 的地址。

### 步骤 2：创建 API Key

1. 在 Elastic Cloud / Kibana 中进入 **Stack Management** -> **API keys**，或在自建集群里创建 Elasticsearch API Key。
2. 这个 Key 只需要有目标日志索引或 data stream 的读取权限，例如 \`logs-*\`。

### 步骤 3：填写 APO 并测试

1. 主流程只需要填写 **名称**、**URL** 和 **API Key**，其中 **URL** 就是 Elasticsearch 的 **Address**。
2. 系统会自动使用默认值：\`authType=apiKey\`、\`logIndexPattern=logs-*\`。
3. 点击 **测试**，通过后再点击 **接入**。
4. 如果要让 AgentSRE 按服务分析日志，Elastic 中的业务日志最好包含常见 Kubernetes 字段，例如 \`kubernetes.pod.name\` / \`k8s.pod.name\` 和 \`kubernetes.namespace\` / \`k8s.namespace.name\`。标准 Elastic Agent、Filebeat 或 OpenTelemetry Kubernetes 日志通常会自动带上这些字段。
5. 如果结果是 warning，表示连接正常，但 APO 暂时无法确认当前日志样本能按服务实例过滤。仍然可以保存接入；等业务日志带有这些字段后，AgentSRE 才能稳定查到具体服务的日志。`,
    'skywalking': `### 配置

此集成仅需要基本的连接参数。只需根据您的环境信息填写以下字段：

1. **地址**：数据源的服务端点或 API 地址（例如 Tempo、Prometheus 或 SkyWalking）。
2. **用户名 / 密码 / Token**（如适用）：访问服务所需的身份验证凭据。

填写完所有字段后，点击 **保存** 以注册集成。`,
    'loki': `### 步骤 1：访问您的 Grafana Cloud 账户

1. 访问 Grafana Cloud 网站并登录您的账户。
2. 点击右上角的 \`My Account\`。 ![](${GrafanaLoki1})
3. 点击您的 Grafana Cloud stack 部分下的 \`Details\`。 ![](${GrafanaLoki2})
4. 点击 \`Loki\` 部分的 \`Details\`。

### 步骤 2：获取您的 Loki 实例详情

1. 向下滚动到 \`Grafana Data Source settings\` 部分。
2. 如果您还没有 \`Password\`，请生成一个。
3. 复制并粘贴以下值到相应的字段：
   - 将 \`URL\` 值输入到 \`Loki Address\` 字段；
   - 将 \`User\` 值输入到 \`Loki Username\` 字段；
   - 将 \`Password\` 值输入到 \`Loki Token\` 字段。
`,
    'grafanaPrometheus': `### 步骤 1：访问您的 Grafana Cloud 账户

1. 访问 Grafana Cloud 网站并登录您的账户。
2. 点击右上角的 \`My Account\`。 ![](${GrafanaTempoCloud1})
3. 点击您的 Grafana Cloud stack 部分下的 \`Details\`。 ![](${GrafanaTempoCloud2})
4. 点击 \`Prometheus\` 部分的 \`Details\`。

### 步骤 2：获取您的 Prometheus 实例详情

1. 向下滚动到 \`Query Endpoint\` 部分。
2. 如果您还没有 \`Password\`，请生成一个。
3. 复制并粘贴以下值到相应的字段：
   - 将 \`Query URL\` 值输入到 \`PrometheusAddress\` 字段；
   - 将 \`Username\` 值输入到 \`Prometheus Username\` 字段；
   - 将 \`Password\` 值输入到 \`Prometheus Token\` 字段。

![](${GrafanaTempoCloud6})`,
}

// Export a function to get the guide map based on language
export const getGuideMdMap = (language: string = 'en') => {
    return language === 'zh' ? guideMdMapZh : guideMdMapEn
}

// Keep backward compatibility
export const guideMdMap = guideMdMapEn
