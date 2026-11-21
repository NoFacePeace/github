# Annotation

Annotation（注解）是写在 `metadata.annotations` 下的字符串键值对，用于附加说明，或向控制器和工具传递约定的配置。

## 1. 给 Service 添加说明

下面的注解记录了 `backend` Service 的用途。

```yaml
metadata:
  name: backend
  annotations:
    example.com/description: "给前端提供 API"
```

## 2. 组件如何使用注解

控制器或工具可以读取它所支持的注解，并按组件的约定执行相应行为。示例中的 `example.com/description` 用于记录说明。按标签匹配对象时，选择器读取的是 `metadata.labels`。
