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
      sha256 ""
    else
      url "https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases/download/v0.1.0/sap-iac_0.1.0_darwin_amd64"
      sha256 ""
    end
  elsif OS.linux?
    if Hardware::CPU.arm?
      url "https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases/download/v0.1.0/sap-iac_0.1.0_linux_arm64"
      sha256 ""
    else
      url "https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases/download/v0.1.0/sap-iac_0.1.0_linux_amd64"
      sha256 ""
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
