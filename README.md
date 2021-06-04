# go-impftermin-notifier

A small go telegram notifier for available __COVID-19__ vaccination.  
Thx to Mark Schmitt for revers engineering!

## build and run the executable

```bash
Usage: ./do
	 lint
	 go-fmt
	 build <arch>      default=amd64 arm,arm64
	 build-container   builds the container image
```

the cli version works like the container version with environment variables

## Dockerfile for K8s

if you like to use the dockerfile within your cluster, you need to set at least the following config:

```yaml
ports:
- name: http
    containerPort: 8080
    protocol: TCP
env:
- name: ZIP_CODE
    value: "<ZIP>"
- name: RADIUS
    value: "<Radius in km>"
- name: DELAY
    value: "<delay in parsable string e.g 180s, 3m>"
- name: TELEGRAM_KEY
    value: "<Telegram Key>"
- name: TELEGRAM_CHAT_ID
    value: "<Telegram Chat ID>"
```