# faas-wasm

## example
First, run the program
```shell
sudo ./main
```

Then use faas-cli to deploy function.
```shell
faas-cli deploy --image your_wasm_file_path --name your_function_name
```

As an example, you can use the wasm file matrix.wasm to have a try.
```shell
faas-cli deploy --image $HOME/go/.../faas-wasm/matrix.wasm --name matrix
```

Please note that you should enter the absolute path of the wasm file.

Then you can invoke this function.
```shell
faas-cli invoke matrix
```
This function is executing a matrix mutiplication.

After finish this function and you don't need it anymore, you can delete it.
```shell
faas-cli delete matrix
```

##Notice
- This platform doesn't support entering parameters for functions. You need to set the write parameter for your function in advance.
- The name of the function you invoke should match the name in the wasm file you write.
- This platform can only run on multi-core system.
