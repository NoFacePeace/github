# smq

```sh
mkdir smq 
cd smq
kubebuilder init --domain localhost --repo localhost/smq
kubebuilder create api --group smq --version v1 --kind Smq
make manifests
make install
make run
```

