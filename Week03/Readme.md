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

## 实践三

拿私有数据进行模型微调

   组装成微调需要的数据格式，上传到 OpenAI 平台，获取数据的file_id,然后指定可微调的模型，创建微调任务，获取微调任务的id

   微调模型需要openapi的key,第三方平台不行,需要申请。后续尝试微调智普的

## 实践四

   [实践代码在这里](./action4/rag运维知识库.ipynb)
   
   RAG运维知识库实践

   数据库 chroma,使用markdown格式的文档数据，根据标题进行切片。使用模型text-embedding-3-small对文档数据进行向量化，存入chroma数据库中。

   然后输入问题，将问题进行向量化，查询数据库中相似度最高的文档，返回文档内容。最后组织prompt，调用chatgpt模型，返回答案。

## 实践五

    (实践失败，环境安装有问题)

   [实践代码在这里](./action5/graphrag.ipynb)

   以增强答案生成的事实性，减少幻觉，便成了自然的选择——RAG(Retrieval-Augmented Generation，检索增强生成)应运而生。

   graphRAG运维知识库实践

## 实践六

  ragflow实践, 开启ragflow需要的环境4核，16G，.env中需要改版本为0.9.0，默认的实践时有问题
  ![](./images/ragflow.png)

## 实践七

  ollama 

  [实践代码在这](./action1/ollama.ipynb)

```
  ollama serve
```

  api_key="ollama",
  base_url="http://localhost:11434/v1",
  
  response = client.chat.completions.create(
    # model="gpt-4o",
    model="qwen2:7b",
    messages=messages,
    tools=tools,
    tool_choice="auto",
)