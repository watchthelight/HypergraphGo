# typed: false
# frozen_string_literal: true

# Homebrew formula for HypergraphGo CLI
# To use with a tap: brew tap watchthelight/tap && brew install hg
class Hg < Formula
  desc "Hypergraph & HoTT tooling in Go"
  homepage "https://github.com/watchthelight/HypergraphGo"
  version "1.3.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v#{version}/hg_#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    else
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v#{version}/hg_#{version}_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v#{version}/hg_#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    else
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v#{version}/hg_#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    bin.install "hg"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/hg -version")
  end
end
