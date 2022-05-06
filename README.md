# kbm-iam

This is an iam project form kbm.


## Run Locally

Clone the project

```bash
  git clone git@github.com:JieTrancender/kbm-iam.git
```

Go to the project directory

```bash
  cd kbm-iam
```

Install dependencies

```bash
  make tidy
```

Start the server

```bash
  make && ./build/platforms/PLATFORM/ARCH/iam-apiserver -c conf/dev.yaml
```
