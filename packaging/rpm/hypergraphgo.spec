# RPM spec for HypergraphGo
# Build: rpmbuild -ba hypergraphgo.spec
# Publish to COPR: https://copr.fedorainfracloud.org

Name:           hypergraphgo
Version:        1.3.0
Release:        1%{?dist}
Summary:        Hypergraph & HoTT tooling in Go

License:        MIT
URL:            https://github.com/watchthelight/HypergraphGo
Source0:        https://github.com/watchthelight/HypergraphGo/releases/download/v%{version}/hg_%{version}_linux_amd64.tar.gz
Source1:        https://github.com/watchthelight/HypergraphGo/releases/download/v%{version}/hg_%{version}_linux_arm64.tar.gz

BuildArch:      x86_64 aarch64
ExclusiveArch:  x86_64 aarch64

%description
A production-quality Go library and CLI for hypergraph theory, supporting
generic vertex types, advanced algorithms, and CLI tools. Includes HoTT
(Homotopy Type Theory) kernel implementation with normalization by evaluation.

%prep
%ifarch x86_64
%setup -q -c -T -a 0
%endif
%ifarch aarch64
%setup -q -c -T -a 1
%endif

%install
install -Dm755 hg %{buildroot}%{_bindir}/hg
install -Dm644 LICENSE.md %{buildroot}%{_licensedir}/%{name}/LICENSE.md
install -Dm644 README.md %{buildroot}%{_docdir}/%{name}/README.md

%files
%{_bindir}/hg
%license %{_licensedir}/%{name}/LICENSE.md
%doc %{_docdir}/%{name}/README.md

%changelog
* Thu Dec 05 2024 watchthelight <admin@watchthelight.org> - 1.3.0-1
- Phase 3: Bidirectional type checking
- macOS DMG releases
- Go 1.25 requirement

* Wed Dec 04 2024 watchthelight <admin@watchthelight.org> - 1.2.0-1
- Phase 2: Normalization and definitional equality
- NbE evaluator with eta-rules
