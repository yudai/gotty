const webpack = require('webpack');
const TerserPlugin = require('terser-webpack-plugin');
const path = require('path');
const licenseLoader = require('./license-loader');

module.exports = {
    mode: "production",
    entry: "./src/main.ts",
    output: {
        path: path.resolve(__dirname, "dist"),
        filename: "gotty-bundle.js"
    },
    devtool: false,
    // xterm.js is large; the bundle is served by GoTTY itself, not over a slow CDN
    performance: { hints: false },
    resolve: {
        extensions: [".ts", ".tsx", ".js"],
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                loader: "ts-loader",
                exclude: /node_modules/
            },
            {
                test: /\.m?js$/,
                include: /node_modules/,
                loader: path.resolve(__dirname, 'license-loader.js')
            }
        ]
    },
    optimization: {
        minimizer: [
            new TerserPlugin({
                extractComments: false,
            }),
        ],
    },
    plugins: [
        // Prepend the licenses collected by license-loader after minification
        new webpack.BannerPlugin({
            banner: () => licenseLoader.licenses.join('\n'),
            raw: true,
            stage: webpack.Compilation.PROCESS_ASSETS_STAGE_REPORT,
        }),
    ],
};
