# typed: false
# frozen_string_literal: true

# Homebrew formula for HypergraphGo CLI
# To use with a tap: brew tap watchthelight/tap && brew install hg
class Hg < Formula
  desc "Hypergraph & HoTT tooling in Go"
  homepage "https://github.com/watchthelight/HypergraphGo"
  version "1.4.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v#{version}/hg_#{version}_darwin_arm64.tar.gz"
      sha256 "634c6c00913fc8fd68c47fedc0446650d35ac07888fc5642ca7085cb69520ab8"
    else
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v#{version}/hg_#{version}_darwin_amd64.tar.gz"
      sha256 "97f0291f5e8376e343036891d05bb83933eff5759934f4662853933b2c89889f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v#{version}/hg_#{version}_linux_arm64.tar.gz"
      sha256 "77d0de2d06dba757e6780fedeb7e0110edf15f871aa4d41a85e97e70f37cb679"
    else
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v#{version}/hg_#{version}_linux_amd64.tar.gz"
      sha256 "e0721e91210785566b808f7ce92246f42e9a2c290e1d87d6ab8f225924eb8ab2"
    end
  end

  def install
    bin.install "hg"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/hg -version")
  end
end
