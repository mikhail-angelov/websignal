module.exports = {
  extends: ['plugin:lit/recommended'],
  parser: 'babel-eslint',
  parserOptions: {
    ecmaVersion: 2018,
    sourceType: 'module',
  },
  env: {
    browser: true,
  },
  plugins: ['html', 'lit'],
  rules: {
    'comma-dangle': 0,
    'no-unused-vars': 'warn',
    'no-unexpected-multiline': 'warn',
  },
}
