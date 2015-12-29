class R < Formula
  desc "r is a contextual, path based, bash history."
  homepage "https://jesselucas.github.io/r/"
  url "https://github.com/jesselucas/r/archive/0.3.2.tar.gz"
  version "0.3.2"
  sha256 "b8acb6d54e4cac80eda17e6946ebccc998dede3388a0c6ddf1c04c0b0e52fa74"

  depends_on "go" => :build

  def install
    ENV["GOPATH"] = buildpath
    ENV["GO15VENDOREXPERIMENT"] = "1"

    (buildpath/"src/github.com/jesselucas/r/").install Dir["*"]
    system "go", "build", "-o", "#{bin}/r", "-v", "github.com/jesselucas/r/"
  end

  def caveats; <<-EOS.undent
    `r` has succesfully installed. Please restart your Bash shell.
    EOS
  end

  test do
    actual = system("r -install")
    expected = true
    actualVersion = pipe_output("#{bin}/r -version")
    assert_block do
      actual == expected
      actualVersion == version
    end
  end
end
