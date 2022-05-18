# nsq-took-kit

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

1. etcd
2. nsq
3. elasticsearch
4. go dependencies
```bash
  make tidy
```

Configure Etcd
```base
  {
    "lookupd-http-addresses":[
        "http://127.0.0.1:4161"
    ],
    "topics":[
        "dev_test"
    ],
    "channel": "nsq_tool_kit",
    "dial-timeout": 6,
    "read-timeout": 60,
    "write-timeout": 6,
    "max-in-flight": 200
}
```

Start the server

```bash
  make && ./build/platforms/PLATFORM/ARCH/nsq-consumer -c conf/nsq-consumer.yaml
```
