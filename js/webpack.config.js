const path = require('path');
const TerserPlugin = require("terser-webpack-plugin");
const LicenseWebpackPlugin = require('license-webpack-plugin').LicenseWebpackPlugin;

module.exports = {
    entry: "./src/main.ts",
    entry: {
        "gotty": "./src/main.ts",
    },
    output: {
        path: path.resolve(__dirname, '../bindata/static/js/'),
    },
    devtool: "source-map",
    resolve: {
        extensions: [".ts", ".tsx", ".js"],
    },
    plugins: [
        new LicenseWebpackPlugin()
    ],
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                loader: "ts-loader",
                exclude: /node_modules/
            },
            {
                test: /\.css$/i,
                use: ["style-loader", "css-loader"],
            },
            {
                test: /\.scss$/i,
                use: ["style-loader", "css-loader", {
                    loader: "sass-loader",
                    options: {
                        sassOptions: {
                            includePaths: ["node_modules/bootstrap/scss"]
                        }
                    }
                }
                ],
            },
        ],
    },
    optimization: {
        minimize: true,
        minimizer: [new TerserPlugin()],
    },
};
