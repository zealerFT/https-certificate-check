# 域名证书检查
Let's Encrypt免费证书用的爽，续期还是有成本，虽然k8s里的ingress可以使用cert-manager来自动续期，但基于oss部署的前端项目等一些特殊部署的项目，不能在自动续期，所以需要时刻关注免费证书的到期时间，
来保证不会遗漏需要手动需求的免费证书。

## 使用
```shell
- make build
- bin/hutao polling || bin/hutao cronjob
- 可增加2个参数
  bin/hutao cronjob --domains=www.baidu.com,www.taobao.com
  bin/hutao polling --polling=3 --domains=www.baidu.com,www.taobao.com
```

## 免费证书
- bin/hutao polling
  轮询有代码层面实现，使用k8s部署，会在部署成功后，每24小时执行检测（轮询时间可设置 --polling）
- bin/hutao cronjob
  使用单次执行，结束则完成。配合K8S的cronjob, 参考`cronjob-v1beta1-hutao.yaml`可更加优雅的调整轮询的时间
