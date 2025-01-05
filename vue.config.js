module.exports = {
  configureWebpack: {
    resolve: {
      fallback: {
        fs: false,
        path: false,
        crypto: false,
      },
    },
  },
};
