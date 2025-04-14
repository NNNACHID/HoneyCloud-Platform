```
██████╗  ██╗███████╗██████╗ ██╗  ██╗ ██████╗ ██╗     ██╗██████╗ ██╗   ██╗███████╗
██╔══██╗███║██╔════╝██╔══██╗██║  ██║██╔═████╗██║    ███║██╔══██╗██║   ██║██╔════╝
██║  ██║╚██║███████╗██████╔╝███████║██║██╔██║██║    ╚██║██║  ██║██║   ██║███████╗
██║  ██║ ██║╚════██║██╔═══╝ ██╔══██║████╔╝██║██║     ██║██║  ██║██║   ██║╚════██║
██████╔╝ ██║███████║██║     ██║  ██║╚██████╔╝███████╗██║██████╔╝╚██████╔╝███████║
╚═════╝  ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝╚═════╝  ╚═════╝ ╚══════╝
```

<img src="docs/honeycloud-dt.png" width=360 height=390>


# 🐝 HoneyCloud-Platform

Automated DevSecOps platform for deploying honeypots in a secure cloud environment based on Kubernetes, Terraform and OpenStack.

## 🚀 Objective

Set up a secure cloud infrastructure capable of dynamically deploying honeypots (decoys), collecting attack data, monitoring activity in real-time, and automating everything through CI/CD pipelines.

---

## 🧱 Tech Stack

| Domain              | Tools used                                       |
|---------------------|--------------------------------------------------|
| Infrastructure      | Terraform, OpenStack                             |
| Orchestration       | Kubernetes (K3s ou MicroK8s)                     |
| Containerization    | Docker                                           |
| Honeypots           | Cowrie, Honeytrap, Go Custom Honeypot            |
| Main language       | Golang                                           |
| CI/CD               | GitHub Actions                                   |
| Secrets Management  | GitHub Secrets, Sealed Secrets                   |
| Monitoring          | Prometheus, Grafana, Loki, Alertmanager          |
| Security            | RBAC, TLS, Trivy, GoSec                          |
| Backup              | Bash scripts, PVCs, CronJobs                     |

---

## 📦 Features

- Auto-deployment of honeypots in a K3s cluster
- Infrastructure as Code (IaC)
- CI/CD with security testing
- Real-time monitoring and alerting
- Logging and attack trace storage
- Supervision dashboard
- Automated backups

---

## 📌 Roadmap

- [x] Specifications
- [ ] Cloud infrastructure via Terraform
- [ ] K3s cluster with TLS ingress
- [ ] Deployment of containerized honeypots
- [ ] GitHub Actions integration (build/test/deploy)
- [ ] Monitoring with Prometheus / Grafana
- [ ] Backup and restoration scripts
- [ ] Final documentation

---


## Project Structure
```
/iac-honeypot/
├── terraform/              # Code IaC (OpenStack, AWS, etc.)
├── kubernetes/
│   ├── honeypots/          # YAML ou Helm Charts pour chaque honeypot
│   └── monitoring/         # Grafana, Prometheus, etc.
├── honeypots/
│   ├── ssh/
│   ├── fake-api-go/
│   └── web-vuln/
├── .github/
│   └── workflows/
│       └── ci-cd.yml       # Pipeline CI/CD GitHub Actions
├── scripts/
│   └── backup.sh
│   └── deploy.sh
├── dashboards/
│   └── grafana/
├── README.md
└── SECURITY.md

```