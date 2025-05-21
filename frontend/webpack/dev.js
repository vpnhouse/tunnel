const path = require('path');
const { merge } = require('webpack-merge');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');

const common = require('./common');
const apiUrl = process.env.npm_config_api ?? 'https://vu-de-2.vpnhouse.net';

printApiBanner(apiUrl);

module.exports = merge(common, {
  mode: 'development',
  stats: 'minimal',
  optimization: {
    moduleIds: 'named',
    removeAvailableModules: false,
    removeEmptyChunks: false,
    splitChunks: false,
  },
  devServer: {
    hot: true,
    historyApiFallback: true,
    port: 8000,
    open: true,
    proxy: {
      '/api': {
        target: apiUrl,
        changeOrigin: true,
      }
    }
  },
  devtool: 'eval-cheap-module-source-map',
  plugins: [
    new ForkTsCheckerWebpackPlugin({
      eslint: {
        enabled: true,
        files: path.resolve(__dirname, '../src')
      },
      issue: {
        scope: "all"
      }
    })
  ]
});


function printApiBanner(apiUrl) {
  const message = `Backend url: ${apiUrl}`;

  console.log('\n'.repeat(2));
  console.log('*'.repeat(message.length));
  console.log(message);
  console.log('*'.repeat(message.length));
  console.log('\n'.repeat(2));
}