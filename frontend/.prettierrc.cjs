module.exports = {
  // 最大行宽
  printWidth: 80,

  // 缩进
  tabWidth: 2,

  // 使用空格而不是制表符
  useTabs: false,

  // 语句末尾的分号
  semi: true,

  // 使用单引号
  singleQuote: true,

  // 对象属性的引号
  quoteProps: 'as-needed',

  // JSX中使用单引号
  jsxSingleQuote: true,

  // 尾随逗号
  trailingComma: 'es5',

  // 对象字面量的大括号空格
  bracketSpacing: true,

  // JSX的右括号位置
  bracketSameLine: false,

  // 箭头函数的括号
  arrowParens: 'always',

  // 文件末尾的换行符
  endOfLine: 'lf',

  // 格式化嵌入代码
  embeddedLanguageFormatting: 'auto',

  // HTML空白处理
  htmlWhitespaceSensitivity: 'css',

  // 插入分号
  insertPragma: false,

  // 需要pragma注释
  requirePragma: false,

  //  prose wrapping
  proseWrap: 'preserve',


  // 排序导入
  importOrder: [
    '^react',
    '^@?\\w', // 第三方库
    '^@/(.*)$', // 内部模块
    '^\\./(?=.*/)(?!.*\\.(css|less|scss|sass)$)', // 同级目录
    '^\\.\\.(?!/?$)', // 上级目录
    '^\\./(?!.*/)(?!.*\\.(css|less|scss|sass)$)', // 当前目录
    '\\.(css|less|scss|sass)$', // 样式文件
  ],
  importOrderSeparation: true,
  importOrderSortSpecifiers: true,
  importOrderCaseInsensitive: true,

  // 插件
  plugins: [
    require.resolve('@trivago/prettier-plugin-sort-imports'),
  ],
};