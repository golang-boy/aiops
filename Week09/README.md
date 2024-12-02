# 预测模型的训练与自动扩容


## 实践一

   1. 部署一个nginx-ingress,流量进入转发到业务pod上
   2. prometheus通过nginx-ingress的metrics获取nginx的流量指标, 
   3. 预测服务从prometheus获取nginx的流量指标,通过模型预测nginx的流量推荐的副本数
   4. operator通过推荐副本数，扩容或缩容pod


流程：

  1. 通过iac部署开启云实例，安装k3s并部署nginx-ingress,同时设置perometheus的监控指标
  2. 开发hpa的operator，通过prometheus获取nginx的流量指标，通过模型预测nginx的流量推荐的副本数
  3. 部署预测服务
     1. 构建镜像
     2. 编写deployment, 并应用
     
