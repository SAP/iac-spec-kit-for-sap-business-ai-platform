# typed: true
# frozen_string_literal: true

# Sap-iac is a formula for installing the Infrastructure-as-Code Specification Toolkit for SAP Business AI Platform
class SapIac < Formula
  desc "Infrastructure-as-Code Specification Toolkit for SAP Business AI Platform"
  homepage "https://sap.github.io/iac-spec-kit-for-sap-business-ai-platform/"
  version "0.1.0"

  if OS.mac?
    if Hardware::CPU.arm?
      url "https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases/download/v0.1.0/sap-iac_0.1.0_darwin_arm64"
      sha256 "1f53347fca6cddd61920dedc9c83b8c89df251ad7a35ce62c87b3ee093a496fa"
    else
      url "https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases/download/v0.1.0/sap-iac_0.1.0_darwin_amd64"
      sha256 "15fb95f06d9cfbfe88df7d15c8b8aa91d8623df6c5e0ef1cf284ef6bb3590914"
    end
  elsif OS.linux?
    if Hardware::CPU.arm?
      url "https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases/download/v0.1.0/sap-iac_0.1.0_linux_arm64"
      sha256 "87bb2a75159ff45d550768ed4ae4be88df9c3c50babc44b27cce317cac40642f"
    else
      url "https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases/download/v0.1.0/sap-iac_0.1.0_linux_amd64"
      sha256 "18695fc9e8dd749ae7d774fd4460aac1a9b75ce8fb04acfbb1d56eaafd4e8b58"
      depends_on arch: :x86_64
    end
  end

  def install
    bin.install stable.url.split("/")[-1] => "sap-iac"
  end

  def caveats
    <<~EOS
      Run:
         sap-iac --help for more information.
    EOS
  end

  test do
     system "#{bin}/sap-iac", "--version"
  end
end
