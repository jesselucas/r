class R < Formula
  desc "r is a contextual, path based, bash history."
  homepage "https://jesselucas.github.io/r/"
  url "https://github.com/jesselucas/r/archive/0.3.3.tar.gz"
  version "0.3.3"
  sha256 "d5eefc25908e56400d7af069d3b979a2f81547f04c2d38e15366f34e467711ff"

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
