第三周
---
https://platform.openai.com/docs/overview

## 实践一

[实践代码在这里](./action1/functionCalling.ipynb)

    实现 Function Calling
        定义 modify_config 函数，入参：service_name，key，value
        定义 restart_service 函数，入参：service_name
        定义 apply_manifest 函数，入参：resource_type，image

## 实践二

[实践代码在这里](./action1/functionCalling.ipynb)

    实践 Function Calling，观察以下输入是否能正确选择对应的函数
        帮我修改 gateway 的配置，vendor 修改为 alipay
        帮我重启 gateway 服务
        帮我部署一个 deployment，镜像是 nginx