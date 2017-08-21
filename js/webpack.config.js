module.exports = {
    entry: "./src/main.ts",
    output: {
        filename: "./dist/bundle.js"
    },
    externals: {
        "hterm": "hterm",
        "htermLib": "lib"
    },
    devtool: "source-map",
    resolve: {
        extensions: [".ts", ".tsx", ".js"],
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                loader: "ts-loader",
                exclude: [/node_modules/],
            }
        ]
    }
};
