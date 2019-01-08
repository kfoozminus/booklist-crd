# K8s Code Generation for CustomResources

### Process
- design types for your CRD
- write artifacts/.yaml to clear structures (not mandatory)
- create these directories
```
root
  artifacts
    *.yaml
  pkg
    apis
      types.go
      doc.go
      register.go
    client
```
- /hack has script files to ease the process of code generation (from [openshift](https://github.com/openshift-evangelists/crd-code-generation))
- run `hack/update-codegen.sh` for code-generation
- run `hack/verify-codegen.sh` to verify if it's up-to-date
- run `go run main.go`
