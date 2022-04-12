const path = require('path');
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ExtractTextPlugin = require('mini-css-extract-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const cssnano = require('cssnano');

require('dotenv').config({ path: path.resolve(__dirname, '../.env') });

process.traceDeprecation = true;
module.exports = {
  entry: [path.resolve(__dirname, '../src/app.tsx')],
  output: {
    path: path.resolve(__dirname, '../dist'),
    publicPath: '/',
    filename: '[name].[contenthash].js'
  },
  module: {
    rules: [
      {
        test: /\.html$/,
        loader: 'html-loader',
        include: path.resolve(__dirname, '../public'),
        exclude: path.resolve(__dirname, '../node_modules'),
        options: {
          minimize: true
        }
      },
      {
        test: /\.(ts|tsx)$/,
        include: path.resolve(__dirname, '../src'),
        exclude: [
          path.resolve(__dirname, '../node_modules'),
          path.resolve(__dirname, '../dist')
        ],
        use: [
          {
            loader: 'ts-loader',
            options: {
              transpileOnly: true,
              experimentalWatchApi: true,
            },
          },
        ]
      },
      {
        test: /\.css$/i,
        include: [
          path.resolve(__dirname, '../src'),
          path.resolve(__dirname, '../node_modules/@fontsource/roboto')
        ],
        use: [
          ExtractTextPlugin.loader,
          {
            loader: 'css-loader',
            options: {
              modules: {
                localIdentName: '[name]__[local]___[hash:base64:5]'
              },
              sourceMap: true,
              importLoaders: 1
            }
          },
          {
            loader: 'postcss-loader',
            options: {
              postcssOptions: {
                ident: 'postcss',
                plugins: [cssnano()]
              }
            }
          },
        ]
      },
      {
        test: /\.(eot|ttf|woff|woff2)$/,
        include: path.resolve(__dirname, '../node_modules/@fontsource/roboto'),
        use: [
          {
            loader: 'file-loader',
            options: {
              name: '[name].[ext]',
              outputPath: 'fonts/'
            }
          }
        ]
      },
      {
        test: /\.(png)$/,
        include: path.resolve(__dirname, '../src'),
        use: [
          {
            loader: 'file-loader',
          }
        ]
      },
      {
        test: /\.svg?$/,
        include: path.resolve(__dirname, '../src'),
        use: [
          {
            loader: 'raw-loader',
          }
        ]
      },
    ]
  },
  resolve: {
    modules: ['node_modules', 'src'],
    extensions: ['index.js', 'index.jsx', 'index.ts', 'index.ts', '.js', '.jsx', '.json', '.ts', '.tsx'],
    alias: {
      '@root': path.resolve(__dirname, '../src/'),
      '@constants': path.resolve(__dirname, '../src/constants'),
      '@common': path.resolve(__dirname, '../src/common'),
      '@context': path.resolve(__dirname, '../src/context'),
      '@apiHooks': path.resolve(__dirname, '../src/apiHooks'),
      '@config': path.resolve(__dirname, '../src/config.json'),
      '@schema': path.resolve(__dirname, '../schema.ts')
    }
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: path.resolve(__dirname, '../public/index.html')
    }),
    new CleanWebpackPlugin(),
    new ExtractTextPlugin({
      filename: '[name].[contenthash].css'
    }),
    new webpack.DefinePlugin({
      "process.env": JSON.stringify(process.env)
    })
  ]
};
