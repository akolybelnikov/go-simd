"""Archive a native simdlab binary and verify the unpacked local UI."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import queue
import subprocess
import tarfile
import tempfile
import threading
import urllib.parse
import urllib.request
import zipfile


def smoke(binary: Path, work: Path) -> None:
    env = {key: value for key, value in os.environ.items() if not key.startswith("GO")}
    env["PATH"] = ""
    with subprocess.Popen(
        [str(binary), "ui", "-no-open"], cwd=work, env=env, stdout=subprocess.PIPE,
        stderr=subprocess.PIPE, text=True, encoding="utf-8",
    ) as process:
        lines = queue.Queue()

        def read_address():
            for line in process.stdout:
                if line.startswith("http://127.0.0.1:"):
                    lines.put(line.strip())
                    return
            lines.put(None)

        threading.Thread(target=read_address, daemon=True).start()
        try:
            url = lines.get(timeout=15)
            if not url:
                raise RuntimeError("The unpacked binary did not print a local UI address.")
            parsed = urllib.parse.urlsplit(url)
            origin = f"{parsed.scheme}://{parsed.netloc}"
            opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))

            def post(path, data=b""):
                request = urllib.request.Request(
                    origin + path, data=data, method="POST",
                    headers={"Origin": origin, "X-Simdlab-Session": parsed.fragment,
                             "Content-Type": "application/json"},
                )
                with opener.open(request, timeout=130) as response:
                    return json.load(response)

            with opener.open(origin + "/", timeout=10) as response:
                if b"Run locally" not in response.read():
                    raise RuntimeError("The unpacked binary did not serve the embedded UI.")
            plan = post("/api/starter")
            if (plan["operation"], plan["size"], plan["repeats"]) != ("Addition", 1024, 3):
                raise RuntimeError("The unpacked binary returned unexpected starter settings.")
            result = post("/api/run", json.dumps({"ref": plan["ref"]}).encode())
            if sum(len(series["repeats"]) for series in result["series"]) != 6:
                raise RuntimeError("The unpacked binary did not retain six repeat measurements.")
            report = Path(result["report_path"])
            measurements = Path(result["evidence"])
            artifacts = work / "artifacts"
            if not report.is_relative_to(artifacts) or report.read_text(encoding="utf-8") != result["report"]:
                raise RuntimeError("The saved Markdown report does not match the browser report.")
            if not measurements.is_relative_to(artifacts) or not measurements.is_file():
                raise RuntimeError("The unpacked binary did not retain the measurement file.")
            opened = post("/api/open", measurements.read_bytes())
            if opened["timing"] != result["timing"]:
                raise RuntimeError("The saved capture did not reproduce the browser result.")
        finally:
            if process.poll() is None:
                process.terminate()
                try:
                    process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait()


def package(binary: Path, platform: str, output: Path) -> None:
    root = Path(__file__).resolve().parents[1]
    quickstart = root / "docs" / "binary-quickstart.txt"
    output.mkdir(parents=True, exist_ok=True)
    stem = f"simdlab-{platform}"
    windows = platform.startswith("windows")
    archive = output / (stem + (".zip" if windows else ".tar.gz"))
    executable = "simdlab.exe" if windows else "simdlab"
    with (
        tempfile.TemporaryDirectory(prefix="simdlab-release-") as tmp,
        tempfile.TemporaryDirectory(prefix=f".{stem}-", dir=output) as staging,
    ):
        work = Path(tmp).resolve()
        unpacked = work / stem / executable
        staged_archive = Path(staging) / archive.name
        files = [(binary, f"{stem}/{executable}"), (quickstart, f"{stem}/README.txt")]
        if windows:
            with zipfile.ZipFile(staged_archive, "w", compression=zipfile.ZIP_DEFLATED) as bundle:
                for source, name in files:
                    bundle.write(source, name)
            with zipfile.ZipFile(staged_archive) as bundle:
                bundle.extractall(work)
        else:
            with tarfile.open(staged_archive, "w:gz") as bundle:
                for source, name in files:
                    bundle.add(source, arcname=name)
            with tarfile.open(staged_archive) as bundle:
                bundle.extractall(work, filter="data")
        smoke(unpacked, work)
        with staged_archive.open("rb") as data:
            checksum = hashlib.file_digest(data, "sha256").hexdigest()
        staged_checksum = Path(staging) / (archive.name + ".sha256")
        staged_checksum.write_text(
            f"{checksum}  {archive.name}\n", encoding="ascii", newline="\n"
        )
        staged_archive.replace(archive)
        staged_checksum.replace(output / staged_checksum.name)
    print(f"Verified {archive.name} with no Go or Ollama setup.")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=Path)
    parser.add_argument("platform", choices=["linux-amd64", "windows-amd64"])
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    package(args.binary.resolve(), args.platform, args.output.resolve())
