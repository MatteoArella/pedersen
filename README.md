# Pedersen's Verifiable Secret Sharing

<p align="center">
    <a href="https://pkg.go.dev/github.com/matteoarella/pedersen?tab=doc"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white" alt="go.dev"></a>
    <a href="https://github.com/matteoarella/pedersen/actions/workflows/main.yml"><img src="https://github.com/matteoarella/pedersen/actions/workflows/main.yml/badge.svg" alt="Build Status"></a>
    <a href="https://github.com/matteoarella/pedersen/releases"><img src="https://img.shields.io/github/release/matteoarella/pedersen.svg" alt="GitHub release"></a>
    <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="Licenses"></a>
</p>

> Secret sharing (also called secret splitting) refers to methods for distributing a secret among a group, in such a way that no individual holds any intelligible information about the secret, but when a sufficient number of individuals combine their 'shares', the secret may be reconstructed.
> [Wikipedia](https://en.wikipedia.org/wiki/Secret_sharing)

Secret sharing schemes are ideal for storing information that is highly sensitive
and highly important, like for example encryption keys. Each of these pieces of
information must be kept highly confidential, as their exposure could be disastrous;
however, it is also critical that they should not be lost. Traditional methods for
encryption are ill-suited for simultaneously achieving high levels of confidentiality
and reliability. This is because when storing the encryption key, one must choose
between keeping a single copy of the key in one location for maximum secrecy,
or keeping multiple copies of the key in different locations for greater reliability.
Increasing reliability of the key by storing multiple copies lowers confidentiality by
creating additional attack vectors. Secret sharing schemes address this problem,
and allow arbitrarily high levels of confidentiality and reliability to be achieved.

> In cryptography, a secret sharing scheme is verifiable if auxiliary information is included that allows players to verify their shares as consistent.
> [Wikipedia](https://en.wikipedia.org/wiki/Verifiable_secret_sharing)

The `pedersen` package implements verifiable secret sharing procedures that are defined by [Pedersen Non-Interactive and Information-Theoretic Secure Verifiable Secret Sharing](https://link.springer.com/chapter/10.1007/3-540-46766-1_9) conference paper.

# Getting started

Any information on how to use this project can be found [here](https://matteoarella.github.io/pedersen).

# License

Pedersen is released under MIT license. See [LICENSE](./LICENSE).
