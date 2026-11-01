class Mutefeed < Formula
  desc "Quiet, clean, AI-slop-free RSS feeds proxy and filter"
  homepage "https://github.com/enlivee/mutefeed"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/enlivee/mutefeed/releases/download/v1.0.0/mutefeed_darwin_arm64.tar.gz"
      # sha256 "..."
    else
      url "https://github.com/enlivee/mutefeed/releases/download/v1.0.0/mutefeed_darwin_amd64.tar.gz"
      # sha256 "..."
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/enlivee/mutefeed/releases/download/v1.0.0/mutefeed_linux_arm64.tar.gz"
      # sha256 "..."
    else
      url "https://github.com/enlivee/mutefeed/releases/download/v1.0.0/mutefeed_linux_amd64.tar.gz"
      # sha256 "..."
    end
  end

  def install
    bin.install "mutefeed"
  end

  service do
    run [opt_bin/"mutefeed", "serve", "--port", "8899"]
    keep_alive true
    log_path var/"log/mutefeed.log"
    error_log_path var/"log/mutefeed.err.log"
  end

  test do
    system "#{bin}/mutefeed", "version"
  end
end
