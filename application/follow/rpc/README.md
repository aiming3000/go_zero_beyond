# followRPC 服务

## 参考文档

https://pwmzlkcu3p.feishu.cn/docx/Si1Cd4EGxoZXkJxGenzcFttOnsh

## 如何结合gorm

go_zero_beyond/pkg/orm/orm.go

go_zero_beyond/pkg/orm/metric.go

go_zero_beyond/pkg/orm/plugin.go

### 修改配置文件

go_zero_beyond/application/follow/rpc/etc/follow.yaml

go_zero_beyond/application/follow/rpc/internal/config/config.go

go_zero_beyond/application/follow/rpc/internal/server/followserver.go

## 编写model层

go_zero_beyond/application/follow/rpc/internal/model/follow.go

go_zero_beyond/application/follow/rpc/internal/model/followcount.go

## 书写业务层


## Prometheus 安装

下载地址：https://prometheus.io/download/
下载tar包，下载后直接解压

[docker安装](https://www.cnblogs.com/Galaxy1/p/18034293)

`docker run  -d --name prometheus --restart=always -p 9090:9090 -p 9101:9101 prom/prometheus`

`docker run  -d --name prometheus --restart=always -p 9090:9090 -v /opt/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus
`

配置文件 #vim prometheus.yml    /etc/prometheus/prometheus.yml
```yaml

scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: 'prometheus'

    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.

    static_configs:
    - targets: ['localhost:9090']

  - job_name: 'file_ds'
    file_sd_configs:
    - files:
      - targets.json
```
target.json配置
```yaml
// target.json配置

[
    {
        "targets": ["127.0.0.1:9101"],
        "labels": {
            "job": "follow",
            "app": "follow",
            "env": "test",
            "instance": "127.0.0.1:9101"
        }
    }
]
```

## Jaeger安装

下载地址：https://www.jaegertracing.io/download/
下载tar包，下载后直接解压

[docker安装](https://www.cnblogs.com/wangzhi8/p/17759461.html)
https://www.cnblogs.com/a120608yby/p/17664696.html

```yaml
docker run --rm --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 14250:14250 \
  -p 14268:14268 \
  -p 14269:14269 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.50



```
最后执行下面命令：

`docker run --name jaeger -d -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one:1.60`

