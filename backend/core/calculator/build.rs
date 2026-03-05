fn main() {
    let protoc = protoc_bin_vendored::protoc_bin_path().expect("failed to fetch protoc path");
    std::env::set_var("PROTOC", protoc);

    tonic_build::configure()
        .build_server(true)
        .build_client(false)
        .compile_protos(&["../../proto/worker.proto"], &["../../proto"])
        .expect("failed to compile worker.proto");
}
