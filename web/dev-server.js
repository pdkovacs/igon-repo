var WebpackDevServer = require('webpack-dev-server');
var webpack = require('webpack');
var config = require('./webpack.config.js');

var restService = 'http://localhost:8090';

var compiler = webpack(config);
var server = new WebpackDevServer(compiler, {
  filename: config.output.filename,
  publicPath: config.output.publicPath,
  contentBase: 'dist/',
  stats: {
    colors: true
  },
  proxy: {
    '/icon': {
      target: restService,
      secure: false
    },
    '/tag': {
      target: restService,
      secure: false
    },
    '/app-info': {
      target: restService,
      secure: false
    },
    '/user': {
      target: restService,
      secure: false
    }
  }
});
server.listen(8080, 'localhost', function () {
});