
# 华为云 Exporter

[华为云](https://www.huaweicloud.com/)云监控的 Prometheus Exporter.

注意：该插件仅适用于华为云局点。

## 介绍
Prometheus是用于展示大型测量数据的开源可视化工具，在工业监控、气象监控、家居自动化和过程管理等领域也有着较广泛的用户基础。将华为云Cloudeye服务接入 prometheus后，您可以利用 prometheus更好地监控和分析来自 Cloudeye服务的数据。

## 拓展标签支持情况
该插件对于已对接云监控的云服务均支持指标数据的导出。为提高云服务资源的识别度、可读性，插件对于以下服务支持导出资源属性label，如ECS实例会增加hostname、ip等label，同时支持将华为云标签转化为label，满足对资源自定义label的诉求，具体如下：
|云服务|命名空间|支持通过实例TAG增加标签|标签数据来源|
| :--- | :--- | :--: | :--: |
|弹性云服务器|SYS.ECS/AGT.ECS|√|RMS/云服务|
|云硬盘|SYS.EVS|√|RMS/云服务|
|分布式缓存服务|SYS.DCS|√|RMS|
|云专线|SYS.DCAAS|√|RMS|
|弹性公网IP和带宽|SYS.VPC|√|RMS|
|云搜索服务|SYS.ES|√|RMS|
|关系型数据库|SYS.RDS|√|RMS|
|弹性负载均衡|SYS.ELB|√|云服务|
|云数据库 GaussDB(for MySQL)|SYS.GAUSSDB|√|RMS|
|云数据库 GaussDB(for openGauss)|SYS.GAUSSDBV5|√|云服务|
|NAT网关|SYS.NAT|√|RMS|
|弹性伸缩|SYS.AS|√|RMS|
|函数工作流|SYS.FunctionGraph|√|RMS|
|数据复制服务|SYS.DRS|√|RMS|
|Web应用防火墙|SYS.WAF|√|RMS|
|文档数据库服务|SYS.DDS|√|云服务|
|API网关|SYS.APIG|×|云服务|
|云备份|SYS.CBR|√|RMS/云服务|
|数据湖探索|SYS.DLI|√|RMS&云服务|
|弹性文件服务|SYS.SFS|×|云服务|
|弹性文件服务 SFS Turbo|SYS.EFS|√|RMS|
|虚拟专用网络|SYS.VPN|√|RMS|
|云数据迁移|SYS.CDM|×|云服务|
|数据仓库服务|SYS.DWS|√|云服务|
|内容审核Moderation|SYS.MODERATION|×|-|
|Anti-DDoS流量清洗|SYS.DDOS|√|RMS|
|云数据库GaussDB(for Nosql)|SYS.NoSQL|×|云服务|
|分布式消息服务|SYS.DMS|√|RMS|
|分布式数据库中间件|SYS.DDMS|×|RMS&云服务|
|API专享版网关|SYS.APIC|×|云服务|
|裸金属服务器|SYS.BMS/SERVICE.BMS|√|RMS|
|ModelArts|SYS.ModelArts|√|RMS|
|VPC终端节点|SYS.VPCEP |√|RMS|
|图引擎服务GES|SYS.GES|√|RMS|
|数据库安全服务DBSS|SYS.DBSS |√|RMS|
|MapReduce服务|SYS.MRS |√|RMS/云服务|
|湖仓构建服务|SYS.LakeFormation |√|RMS/云服务|
|智能数据湖运营平台|SYS.DAYU |√|云服务|
|云防火墙|SYS.CFW |√|RMS|
|广域网质量监控|SYS.WANQMonitor |√|云服务|
|云日志服务|SYS.LTS |×|云服务|
|视频直播|SYS.LIVE |×|云服务|
|云原生应用网络|SYS.ANC|√|RMS|
|企业主机安全|SYS.HSS|×|云服务|
|表格存储服务|SYS.CloudTable|×|云服务|
|事件网格|SYS.EG|√|RMS|
|对象存储服务|SYS.OBS|√|RMS|
|云解析服务|SYS.DNS|√|云服务|
|企业门户|SYS.EWP|√|云服务|

注：自定义标签时，key只能包含大写字母、小写字母以及中划线

## 环境准备
以Ubuntu 18.04系统和Prometheus 2.14.0版本为例
| Prometheus | prometheus-2.14.0.linux-amd64 |
| ------------ | ------------ |
| ECS | Ubuntu 18.04 |
| Ubuntu private ip | 192.168.0.xx |

账号要求具有IAM，CES，RMS，EPS服务的可读权限,另外获取哪些服务的监控数据就需要有哪些服务的只读权限

## 安装配置cloudeye-exporter
1. 在ubuntu vm上安装cloudeye-exporter

   登录vm机器，查看插件Releases版本 (https://github.com/huaweicloud/cloudeye-exporter/releases) ，获取插件下载地址，下载解压安装。
```
# 参考命令：
mkdir cloudeye-exporter
cd cloudeye-exporter
wget https://github.com/huaweicloud/cloudeye-exporter/releases/download/v2.0.5/cloudeye-exporter.v2.0.5.tar.gz
tar -xzvf cloudeye-exporter.v2.0.5.tar.gz
```
2. 编辑clouds.yml文件配置公有云信息

   区域ID以及auth_url可点击下面链接查看
 *  [地区和终端节点（中国站）](https://developer.huaweicloud.com/endpoint?IAM)
 *  [地区和终端节点（国际站）](https://developer.huaweicloud.com/intl/en-us/endpoint?IAM)
```
global:
  port: "{private IP}:8087" # 监听端口 :出于安全考虑，建议不将expoter服务端口暴露到公网，建议配置为127.0.0.1:{port}，或{内网ip}:{port}，例如：192.168.1.100:8087；如业务需要将该端口暴露到公网，请确保合理配置安全组，防火墙，iptables等访问控制策略，确保最小访问权限
  scrape_batch_size: 300
  resource_sync_interval_minutes: 20 # 资源信息更新频率：默认180分钟更新一次；该配置值小于10分钟，将以10分钟1次为资源信息更新频率
  ep_ids: "xxx1,xxx2" # 可选配置，根据企业项目ID过滤资源，不配置默认查询所有资源的指标，多个ID使用英文逗号进行分割。
  logs_conf_path: "/root/logs.yml" # 可选配置，指定日志打印配置文件路径，建议使用绝对路径。若未指定，程序将默认使用执行启动命令所在目录下的日志配置文件。
  metrics_conf_path: "/root/metric.yml" # 可选配置，指定指标配置文件路径，建议使用绝对路径。若未指定，程序将默认使用执行启动命令所在目录下的指标配置文件。
  endpoints_conf_path: "/root/endpoints.yml" # 可选配置，指定服务域名配置文件路径，建议使用绝对路径。若未指定，程序将默认使用执行启动命令所在目录下的服务域名配置文件。
  ignore_ssl_verify: false # 可选配置，exporter查询资源/指标时默认校验ssl证书；若用户因ssl证书校验导致功能异常，可将该配置项配置为true跳过ssl证书校验
  client_cn: "" # 可选配置，指定被信任的DNS列表，用于校验https证书链中的DNS名称，列表中多个DNS名称请以逗号分隔
  
  # 可选配置，根据实际情况替换示例中的代理协议、地址和端口号
  proxy_schema: "http"                      
  proxy_host: "proxy.huaweicloud.com"        
  proxy_port: 8080          
  # 可选配置，如果代理需要认证，请配置用户名和密码
  proxy_username: "username"                 
  proxy_password: "password"                 
auth:
  auth_url: "https://iam.{region_id}.myhuaweicloud.com/v3"
  project_name: "cn-north-1" # 华为云项目名称，可以在“华为云->统一身份认证服务->项目”中查看
  access_key: "" # IAM用户访问密钥 您可参考4.1章节，使用脚本将ak sk解密后传入，避免因在配置文件中明文配置AK SK而引发信息泄露
  secret_key: ""
  region: "cn-north-1" # 区域ID
```
注：默认的监控端口为8087.

3. 欧洲站注意事项

   欧洲站用户需要为rms、eps服务重新指定域名，如下所示
```yaml
"rms":
  "https://rms.myhuaweicloud.eu"
"eps":
  "https://eps.eu-west-101.myhuaweicloud.eu"
```
注：其他站点用户无需在此文件配置域名信息

4. 启动cloudeye-exporter，默认读取当前目录下的clouds.yml文件，也可使用-config参数指定clouds.yml文件路径
```
./cloudeye-exporter -config=clouds.yml
```

4.1 出于安全考虑cloudeye-exporter提供了-s参数, 可以通过命令行交互的方式输入ak sk避免明文配置在clouds.yml文件中引起泄露。
```shell
./cloudeye-exporter -s true
```
下面是shell脚本启动的样例，建议在脚本中配置加密后的ak&sk，并通过您自己的解密方法对ak sk进行解密后通过huaweiCloud_AK和huaweiCloud_SK参数传入cloudeye-exporter。
```shell
#!/bin/bash
## 为了防止您的ak&sk泄露，不建议在脚本中配置明文的ak sk
huaweiCloud_AK=your_decrypt_function("加密的AK")
huaweiCloud_SK=your_decrypt_function("加密的SK")
./cloudeye-exporter -s true<<EOF
$huaweiCloud_AK $huaweiCloud_SK
EOF
```
4.2 出于安全考虑，cloudeye-exporter提供了-k参数，该参数代表以https协议（双向认证）启动本服务；启动前请确保操作系统已安装openssl，启动时需要通过命令行交互方式输入CA证书路径、服务端https证书路径、服务端https私钥路径以及私钥密码（建议客户使用自己公司购买的通过权威认证的CA证书制作双向证书，也支持自签名证书）。
```shell
./cloudeye-exporter -k true
```
下面是shell脚本启动的样例，建议在脚本中配置加密后的CA证书、服务端https证书、私钥以及私钥密码信息，以参数形式传入cloudeye-exporter，避免明文泄露
```shell
#!/bin/bash
huaweiCloud_https_ca=your_decrypt_function("加密的CA证书")
huaweiCloud_https_crt=your_decrypt_function("加密的https证书")
huaweiCloud_https_key=your_decrypt_function("加密的https私钥")
huaweiCloud_https_password=your_decrypt_function("加密的https私钥密码")
root_path="/opt/cloud/cloudeye-exporter/root_crt"
server_crt_path="/opt/cloud/cloudeye-exporter/server_crt"
server_key_path="/opt/cloud/cloudeye-exporter/server_key"
# CA证书写入CA临时文件，exporter读取完毕后会自动删除
cat > $root_path << EOF
$huaweiCloud_https_ca
EOF
# https证书写入证书临时文件，exporter读取完毕后会自动删除
cat > $server_crt_path << EOF
$huaweiCloud_https_crt
EOF
# 私钥写入私钥临时文件，exporter读取完毕后会自动删除
cat > $server_key_path << EOF
$huaweiCloud_https_key
EOF
# 请不要将原始证书路径和私钥路径作为exporter启动参数，以免被exporter删除
./cloudeye-exporter -k true<<EOF
$root_path $server_crt_path $server_key_path $huaweiCloud_https_password
EOF
```
4.3 出于安全考虑cloudeye-exporter提供了-p参数, 可以通过命令行交互的方式输入代理的userName和password，避免明文配置在clouds.yml文件中引起泄露。
```shell
./cloudeye-exporter -p true
```
下面是shell脚本启动的样例，建议在脚本中配置proxyUserName和加密后的proxyPassword，并通过您自己的解密方法对proxyPassword
进行解密，将proxyUserName和解密后的proxyPassword以参数方式传入cloudeye-exporter。
```shell
#!/bin/bash
## 为了防止您的代理泄露，不建议在脚本中配置明文的proxyPassword
huaweiCloud_ProxyUserName="你的userName"
huaweiCloud_ProxyPassword=your_decrypt_function("加密的proxyPassword")
./cloudeye-exporter -p true<<EOF
$huaweiCloud_ProxyUserName $huaweiCloud_ProxyPassword
EOF
```

4.4 在多参数场景下，即ak/sk、proxy信息和https信息都需要通过命令行参数传入，则顺序为-s -p -k。
下面是shell脚本启动的样例，建议在脚本中配置加密后的ak、sk、proxyPassword和CA证书、
服务端https证书、私钥以及私钥密码信息以及proxyUserName信息，并分别以huaweiCloud_AK、huaweiCloud_SK、huaweiCloud_ProxyUserName、
huaweiCloud_ProxyPassword、root_path、server_crt_path、server_key_path、
huaweiCloud_https_password参数传入cloudeye-exporter。
```shell
#!/bin/bash
huaweiCloud_AK=your_decrypt_function("加密的AK")
huaweiCloud_SK=your_decrypt_function("加密的SK")

huaweiCloud_ProxyUserName="你的userName"
huaweiCloud_ProxyPassword=your_decrypt_function("加密的proxyPassword")

huaweiCloud_https_ca=your_decrypt_function("加密的CA证书")
huaweiCloud_https_crt=your_decrypt_function("加密的https证书")
huaweiCloud_https_key=your_decrypt_function("加密的https私钥")
huaweiCloud_https_password=your_decrypt_function("加密的https私钥密码")

root_path="/opt/cloud/cloudeye-exporter/root_crt"
server_crt_path="/opt/cloud/cloudeye-exporter/server_crt"
server_key_path="/opt/cloud/cloudeye-exporter/server_key"

# CA证书写入CA临时文件，exporter读取完毕后会自动删除
cat > $root_path << EOF
$huaweiCloud_https_ca
EOF
# https证书写入证书临时文件，exporter读取完毕后会自动删除
cat > $server_crt_path << EOF
$huaweiCloud_https_crt
EOF
# 私钥写入私钥临时文件，exporter读取完毕后会自动删除
cat > $server_key_path << EOF
$huaweiCloud_https_key
EOF
./cloudeye-exporter -s=true -p=true -k=true<<EOF
$huaweiCloud_AK $huaweiCloud_SK
$huaweiCloud_ProxyUserName $huaweiCloud_ProxyPassword
$root_path $server_crt_path $server_key_path $huaweiCloud_https_password
EOF
```

5. 指标配置
cloudeye-exporter支持的指标可在 [支持监控的服务列表](https://support.huaweicloud.com/usermanual-ces/zh-cn_topic_0202622212.html)中查看。若发现您需要的指标没有获取到，可查阅该文档将对应的指标填入metric.yml中。
   


## 安装配置prometheus接入cloudeye
1. 下载Prometheus (https://prometheus.io/download/)
```
$ wget https://github.com/prometheus/prometheus/releases/download/v2.14.0/prometheus-2.14.0.linux-amd64.tar.gz 
$ tar xzf prometheus-2.14.0.linux-amd64.tar.gz
$ cd prometheus-2.14.0.linux-amd64
```
2. 配置接入cloudeye exporter节点

   修改prometheus中的prometheus.yml文件配置。如下配置所示在scrape_configs下新增job_name名为’huaweicloud’的节点。其中targets中配置的是访问cloudeye-exporter服务的ip地址和端口号，services配置的是你想要监控的服务，比如SYS.VPC,SYS.RDS。
   ```
   $ vi prometheus.yml
   global:
     scrape_interval: 1m # 设置prometheus从exporter查询数据的间隔时间，prometheus配置文件中默认为15s，建议设置为1m
     scrape_timeout: 1m # 设置从exporter查询数据的超时时间，prometheus配置文件中默认为15s，建议设置为1m
   scrape_configs:
     - job_name: 'huaweicloud'
       scheme: https # 默认该值为http，代表prometheus以http方式调用cloudeye-exporter；如需以https方式调用，请将配置值改为https
       tls_config:   # prometheus以https调用cloudeye-exporter服务时，该参数必配
         ca_file: root.crt     # 客户端CA证书的路径
         cert_file: client.crt # 客户端https证书的路径
         key_file: client.key  # 客户端私钥的路径
       static_configs:
       - targets: ['192.168.0.xx:8087'] # exporter节点地址:监听端口
       params:
         services: ['SYS.VPC,SYS.RDS'] # 当前任务需要查询的服务命名空间，建议每个服务配置单独job
   ```
3. 启动prometheus监控华为云服务
```
./prometheus
```
* 登录http://127.0.0.1:9090/graph
* 查看指定指标的监控结果

## Grafana监控面板使用
Grafana是一个开源的可视化和分析平台，支持多种数据源，提供多种面板、插件来快速将复杂的数据转换为漂亮的图形和可视化的工具。将华为云Cloudeye服务接入 prometheus后，您可以利用Grafana更好地分析和展示来自Cloudeye服务的数据。
目前提供了ECS等服务的监控面板，具体使用方法见：[Grafana监控面板使用](./grafana_dashboard/use_grafana_template.md)