module.exports = {
	target: ['web', 'browserslist:IE 8'],
	optimization: {
		minimize: false
	},
	module: {
		rules: [
			{
				test: /\.(js)$/,
				exclude: /node_modules/,
				use: {
					loader: 'babel-loader',
					options:{
						presets: [
								[
										"@babel/preset-env",
										{
												modules: false,
												forceAllTransforms: true,
												useBuiltIns: "usage",
												corejs: 3,
												targets: {
														browsers: "> 0.5%, last 5 versions",
														chrome: "41",
														firefox: "41",
														safari: "9",
														ios: "7",
														android: "4.1",
														ie: "8"
												}
										}
								]
						],
						plugins: [
								"@babel/plugin-transform-block-scoping",
								"@babel/plugin-transform-runtime"
						]
					}
				}
			},
			{
				test: () => { return true },
				sideEffects: true,
			}
		]
	},
	resolve: {
		extensions: ['*', '.js']
	},
	output: {
		filename: 'main.js'
	}
};
