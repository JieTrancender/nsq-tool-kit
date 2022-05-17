# kbm-iam

This is a nsq tool kit project, can consume nsq messages and post to others.


## Run Locally

Clone the project

```bash
  git clone git@github.com:JieTrancender/nsq-tool-kit.git
```

Go to the project directory

```bash
  cd nsq-tool-kit
```

Install dependencies

```bash
  make tidy
```

Start the server

```bash
  make && ./build/platforms/PLATFORM/ARCH/nsq-consumer -c conf/nsq-consumer.yaml
```
