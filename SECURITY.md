# Security

Gauge-Keeper-Go is a research build. It has not been audited and there is no bug bounty.

## Reporting

Please report a vulnerability privately by email to adiiiresearch@gmail.com or biradaraditya9712@gmail.com rather than in a public issue. Include the affected file or package, the steps to reproduce, and the impact you expect.

## Scope

The keeper in this repository: the allocation logic in `internal/brain`, the bounds check in `internal/validate`, the risk and yield readers in `internal/risk`, and the signing path in `internal/execute`. The Vault Router contracts and the external protocols the keeper reads are out of scope here.
