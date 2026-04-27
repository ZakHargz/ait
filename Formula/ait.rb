# typed: false
# frozen_string_literal: true

class Ait < Formula
  desc "AI Toolkit Package Manager - npm for AI agents, skills, and prompts"
  homepage "https://github.com/ZakHargz/ait"
  url "https://github.com/ZakHargz/ait/archive/refs/tags/v0.9.0.tar.gz"
  sha256 "3f2a2e8f8580d452b8b4781bc7f3dd2f54d7b430b9d31ca3bdbfb449862d4717"
  license "MIT"
  head "https://github.com/ZakHargz/ait.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/ait"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/ait --version")
    
    # Test init command
    system bin/"ait", "init", "--defaults"
    assert_predicate testpath/"ait.yml", :exist?
    
    # Verify ait.yml content
    ait_yml = File.read(testpath/"ait.yml")
    assert_match "name:", ait_yml
    assert_match "version:", ait_yml
    
    # Test doctor command
    system bin/"ait", "doctor"
  end
end
