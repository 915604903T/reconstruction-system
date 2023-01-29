#!/bin/bash

#faas-cli deploy --name deriche --image /home/nano/PolyBenchC-4.2.1/build/wasmstatic/deriche.wasm
faas-cli deploy --name gemm --image /home/nano/PolyBenchC-4.2.1/build/wasmstatic/gemm.wasm
faas-cli deploy --name symm --image /home/nano/PolyBenchC-4.2.1/build/wasmstatic/symm.wasm
faas-cli deploy --name trmm --image /home/nano/PolyBenchC-4.2.1/build/wasmstatic/trmm.wasm
faas-cli deploy --name syrk --image /home/nano/PolyBenchC-4.2.1/build/wasmstatic/syrk.wasm
# faas-cli deploy --name correlation --image /home/nano/PolyBenchC-4.2.1/build/wasmstatic/correlation.wasm
# faas-cli deploy --name covariance --image /home/nano/PolyBenchC-4.2.1/build/wasmstatic/covariance.wasm
faas-cli deploy --name doigten --image /home/nano/PolyBenchC-4.2.1/build/wasmstatic/doitgen.wasm
