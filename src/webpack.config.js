const path = require('path');
require('babel-polyfill');

module.exports = {
  context: path.join(__dirname, './scripts'),
  entry: ['babel-polyfill', './app.jsx'],
  output: {
    path: path.join(__dirname, '../static/'),
    publicPath: '/',
    filename: 'build.js',
  },
  module: {
    rules: [
      {
        test: /\.jsx?$/,
        exclude: /(node_modules)/,
        use: {
          loader: 'babel-loader',
          options: {
            presets: ['@babel/preset-react', '@babel/preset-env'],
          },
        },
      },
      {
        test: /\.s?css$/,
        loaders: ['style-loader', 'css-loader?modules'],
      },
      {
        test: /\.(jpe?g|png|gif|svg|ico)(\?.+)?$/,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: 1024,
              name: './img/[name].[ext]',
            },
          },
        ],
      },
    ],
  },
  plugins: [],
};
