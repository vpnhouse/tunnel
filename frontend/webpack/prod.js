const { merge } = require('webpack-merge');
const CompressionPlugin = require('compression-webpack-plugin');
const common = require('./common');

module.exports = merge(common, {
  mode: 'production',
  plugins: [
    new CompressionPlugin({
      filename: '[path][base].gz[query]',
      algorithm: 'gzip',
      test: /\.js$|\.css$|\.jsx|\.tsx|\.ts|\.svg|\.html$/,
      threshold: 10240,
      minRatio: 0.8
    })
  ]
});
