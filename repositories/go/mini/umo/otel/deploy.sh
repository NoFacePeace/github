#!/bin/bash
# 使用 bash 执行；首行 shebang 供直接 ./deploy.sh 时由内核选用解释器

# 脚本所在目录的绝对路径（readlink -f 解析软链；GNU/Linux 常见，macOS 默认 readlink 无 -f）
WORKDIR=`dirname $(readlink -f $0)`

env=""
eks=""

# 解析命令行参数
while getopts "e:k:" arg
do
    case $arg in
        e)
        env=$OPTARG
        echo "set env, arg:$env"
        ;;
        k)
        eks=$OPTARG
        echo "set eks, arg:$eks"
        ;;
    esac
done

# 检查必需参数
if [ "$env" = "" ]; then
    echo "need to set env with -e option"
    exit 1
fi
if [ "$eks" = "" ]; then
    echo "need to set eks with -k  option"
    exit 1
fi

ns=$(kubectl get namespace)
if [ $? -ne 0 ]; then
    echo "get namespace error"
    exit 1
fi
echo -e "$ns"

echo -e "$ns" | grep -w "cert-manager"
if [ $? -ne 0 ]; then
    echo "need to install cert-manager"
    kubectl apply -f cert-manager.yaml
    if [ $? -ne 0 ]; then
        echo "install cert-manager error"
        exit 1
    fi
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=300s
    if [ $? -ne 0 ]; then
        echo "wait cert-manager error"
        exit 1
    fi
    exit 1
fi

echo "cert-manager installed"

# install opentelemetry-operator
# kubectl apply -f opentelemetry-operator.yaml


# check if opentelemetry-operator-selfsigned-issuer is installed
kubectl get issuers.cert-manager.io -n opentelemetry-operator-system | grep opentelemetry-operator-selfsigned-issuer
if [ $? -ne 0 ]; then
    exit 1
fi



crd=$(kubectl get crd)
if [ $? -ne 0 ]; then
    echo "get cluster crd error"
    exit 0
fi

echo -e "$crd" | grep -w "servicemonitors.monitoring.coreos.com"
if [ $? -ne 0 ]; then
    echo "need to install servicemonitors.yaml"
    kubectl apply -f servicemonitors.yaml
    if [ $? -ne 0 ]; then
        echo "install servicemonitors.yaml failed"
        exit 1
    fi

    echo "install servicemonitors.yaml finish"
fi
echo "servicemonitors.yaml setup finish"

echo -e "$crd" | grep -w "podmonitors.monitoring.coreos.com"
if [ $? -ne 0 ]; then
    echo "need to install podmonitors.yaml"
    kubectl apply -f podmonitors.yaml

    if [ $? -ne 0 ]; then
        echo "install podmonitors.yaml failed"
        exit 1
    fi

    echo "install podmonitors.yaml finish"
fi
echo "podmonitors.yaml setup finish"

kubectl  create namespace opentelemetry-collectors

kubectl apply -f targetallocator_rbac.yaml
if [ $? -ne 0 ]; then
    exit 1
fi


if [ "$env" = "test" ]; then
    sed  -i 's\insert/1/prometheus\insert/2/prometheus\g' targetallocator_and_collector.yaml 
fi


kubectl apply -f targetallocator_and_collector.yaml
if [ $? -ne 0 ]; then
    exit 1
fi

kubectl apply -f podmonitor.yaml
if [ $? -ne 0 ]; then
    exit 1
fi