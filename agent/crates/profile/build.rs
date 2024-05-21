use std::env;
use std::ffi::OsStr;
use std::path::PathBuf;

use libbpf_cargo::SkeletonBuilder;

const BPF_PATH: &str = "src/bpf";
const SRC: &str = "python";

fn main() {
    let mut base_path = PathBuf::from(env::var_os("CARGO_MANIFEST_DIR").unwrap());
    base_path.extend(BPF_PATH.split("/"));

    let mut include_path = base_path.clone();
    include_path.push("include");

    let mut path = base_path.clone();
    path.push(SRC);

    let mut src_path = path.clone();
    src_path.set_extension("bpf.c");
    path.set_extension("skel.rs");
    SkeletonBuilder::new()
        .source(&src_path)
        .clang_args([OsStr::new("-I"), OsStr::new(&include_path)])
        .build_and_generate(&path)
        .unwrap();
    println!("cargo:rerun-if-changed={}", src_path.display());
}
