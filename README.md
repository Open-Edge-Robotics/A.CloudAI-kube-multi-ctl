# A.CloudAI-kube-multi-ctl
Customized kubectl to manage multiple k8s master node (standalone node)

```
연구개발과제명: 일상생활 공간에서 자율행동체의 복합작업 성공률 향상을 위한 자율행동체 엣지 AI SW 기술 개발

세부 개발 카테고리
● 지속적 지능 고도화를 위한 자율적 흐름제어 학습 프레임워크 기술 분석 및 설계
- 기밀성 데이터 활용 지능 고도화를 위한 엣지와 클라우드 분산 협업 학습 프레임워크 기술
- 엣지와 클라우드 협력 학습 간 최적 자원 활용 및 지속적 지능 배포를 위한 자율적 학습흐름제어 기술

개발 내용 
- 엣지와 클라우드 분산 협업을 위한 지속적 지능 배포 프레임워크 
- 자율행동체 엣지 기반 클러스터링 솔루션 및 분산 학습 프레임워크 개발
```

# Multi Deploy to AMR robots Application

This application is used to deploy applications to AMR robots.

## Prerequisites

- [Go](https://go.dev) 1.23 or later
- [K3s](https://k3s.io) Kubernetes cluster

## Build and Deploy

[Build Guide](kmctl/README.md) for more information.

## Quick Start

### Client Installation

```bash
# Client Install
cd kmctl
make install
```

### Client Commands

```bash
# 로봇 클러스터 목록 조회
kmctl get nodes

# 로봇 클러스터 node 상세 조회
kmctl get node -n <node-name>
kmctl get node -n <node-name> -s <node-namespace>

# 로봇 클러스터 Pod 목록 조회
kmctl get pods
kmctl get pods -s <node-namespace>

# 로봇 클러스터 Pod 상세 조회
kmctl get pod -n <pod-name>
kmctl get pod -n <pod-name> -s <node-namespace>

# 로봇 클러스터 Log 조회
kmctl logs -n <pod-name>
kmctl logs -n <pod-name> -s <node-namespace>

# 로봇 클러스터 배포
kmctl apply -f <yaml-file>

# 로봇 클러스터 배포 취소
kmctl delete -f <yaml-file>
```

### Edit Config

```bash
vi ~/.config/kmctl/config.yaml
```

```yaml
# config.yaml
# 배포 대상 로봇 클러스터 정보 등록
server:
- name: "server1"
  host: "192.168.50.11"
  port: 30300
- name: "server2"
  host: "192.168.50.12"
  port: 30300
- name: "server3"
  host: "192.168.50.13"
  port: 30300
# ... 추가 가능
```
