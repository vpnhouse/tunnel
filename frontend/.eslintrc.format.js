const path = require('path');

module.exports = {
  parser: "@typescript-eslint/parser",
  parserOptions: {
    project: "./tsconfig.json",
    sourceType: "module",
    ecmaVersion: 6
  },
  plugins: [
    'react',
    "@typescript-eslint",
    "import"
  ],
  env: {
    browser: true,
    node: true,
    jest: true,
    es6: true
  },
  settings: {
    'import/resolver': {
      node: {
        extensions: ['.mjs', '.js', '.jsx', 'ts', 'tsx', '.json', '.scss', '.css']
      },
      webpack: {
        config: path.join(__dirname, './webpack/common.js')
      },
      typescript: {},
    },
    'import/extensions': ['.js', '.mjs', '.jsx', 'ts', 'tsx'],
    'import/ignore': ['node_modules', '\\.(coffee|scss|css|less|hbs|svg|json)$'],
    react: {
      pragma: 'React',
      version: '16.0'
    },
    propWrapperFunctions: ['forbidExtraProps', 'exact', 'Object.freeze'],
  },
  rules: {
    'no-extra-boolean-cast': 'error',
    'no-extra-parens': [
      'off',
      'all',
      {
        conditionalAssign: true,
        nestedBinaryExpressions: false,
        returnAssign: false,
        enforceForArrowConditionals: false
      }
    ],
    'no-extra-semi': 'error',
    'no-regex-spaces': 'error',
    'no-unsafe-negation': 'error',
    'curly': ['error', 'multi-line'],
    'dot-location': ['error', 'property'],
    'dot-notation': [
      'error',
      {
        allowKeywords: true
      }
    ],
    'eqeqeq': [
      'error',
      'always',
      {
        null: 'ignore'
      }
    ],
    'no-else-return': [
      'error',
      {
        allowElseIf: false
      }
    ],
    'no-extra-bind': 'error',
    'no-extra-label': 'error',
    'no-floating-decimal': 'error',
    'no-implicit-coercion': [
      'off',
      {
        boolean: false,
        number: true,
        string: true
      }
    ],
    'no-multi-spaces': [
      'error',
      {
        ignoreEOLComments: false
      }
    ],
    'no-unused-labels': 'error',
    'no-useless-return': 'error',
    'wrap-iife': [
      'error',
      'outside',
      {
        functionPrototypeMethods: false
      }
    ],
    'yoda': 'error',
    'no-undef-init': 'error',
    'array-bracket-newline': ['off', 'consistent'],
    'array-bracket-spacing': ['error', 'never'],
    'array-element-newline': [
      'off',
      {
        multiline: true,
        minItems: 3
      }
    ],
    'block-spacing': ['error', 'always'],
    'brace-style': [
      'error',
      '1tbs',
      {
        allowSingleLine: true
      }
    ],
    'capitalized-comments': [
      'off',
      'never',
      {
        line: {
          ignorePattern: '.*',
          ignoreInlineComments: true,
          ignoreConsecutiveComments: true
        },
        block: {
          ignorePattern: '.*',
          ignoreInlineComments: true,
          ignoreConsecutiveComments: true
        }
      }
    ],
    'comma-dangle': [
      'error',
      {
        arrays: 'never',
        objects: 'never',
        imports: 'never',
        exports: 'never',
        functions: 'never'
      }
    ],
    'comma-spacing': [
      'error',
      {
        before: false,
        after: true
      }
    ],
    'comma-style': [
      'error',
      'last',
      {
        exceptions: {
          ArrayExpression: false,
          ArrayPattern: false,
          ArrowFunctionExpression: false,
          CallExpression: false,
          FunctionDeclaration: false,
          FunctionExpression: false,
          ImportDeclaration: false,
          ObjectExpression: false,
          ObjectPattern: false,
          VariableDeclaration: false,
          NewExpression: false
        }
      }
    ],
    'computed-property-spacing': ['error', 'never'],
    'eol-last': ['error', 'always'],
    'func-call-spacing': ['error', 'never'],
    'function-paren-newline': ['error', 'consistent'],
    indent: [
      'error',
      2,
      {
        SwitchCase: 1,
        VariableDeclarator: 1,
        outerIIFEBody: 1,
        FunctionDeclaration: {
          parameters: 1,
          body: 1
        },
        FunctionExpression: {
          parameters: 1,
          body: 1
        },
        CallExpression: {
          arguments: 1
        },
        ArrayExpression: 1,
        ObjectExpression: 1,
        ImportDeclaration: 1,
        flatTernaryExpressions: false,
        ignoredNodes: [
          'JSXElement',
          'JSXElement > *',
          'JSXAttribute',
          'JSXIdentifier',
          'JSXNamespacedName',
          'JSXMemberExpression',
          'JSXSpreadAttribute',
          'JSXExpressionContainer',
          'JSXOpeningElement',
          'JSXClosingElement',
          'JSXText',
          'JSXEmptyExpression',
          'JSXSpreadChild'
        ],
        ignoreComments: false
      }
    ],
    'jsx-quotes': ['error', 'prefer-double'],
    'key-spacing': [
      'error',
      {
        beforeColon: false,
        afterColon: true
      }
    ],
    'keyword-spacing': [
      'error',
      {
        before: true,
        after: true,
        overrides: {
          return: {
            after: true
          },
          throw: {
            after: true
          },
          case: {
            after: true
          }
        }
      }
    ],
    'linebreak-style': ['error', 'unix'],
    'lines-around-comment': 'off',
    'lines-between-class-members': [
      'error',
      'always',
      {
        exceptAfterSingleLine: false
      }
    ],
    'multiline-comment-style': ['off', 'starred-block'],
    'new-parens': 'error',
    'newline-per-chained-call': [
      'error',
      {
        ignoreChainWithDepth: 4
      }
    ],
    'no-lonely-if': 'error',
    'no-multiple-empty-lines': [
      'error',
      {
        max: 2,
        maxEOF: 0
      }
    ],
    'no-trailing-spaces': [
      'error',
      {
        skipBlankLines: false,
        ignoreComments: false
      }
    ],
    'no-unneeded-ternary': [
      'error',
      {
        defaultAssignment: false
      }
    ],
    'no-whitespace-before-property': 'error',
    'nonblock-statement-body-position': ['error', 'beside'],
    'object-property-newline': ['error', { allowAllPropertiesOnSameLine: false }],
    'object-curly-newline': [
      'error',
      {
        ObjectExpression: {
          minProperties: 4,
          multiline: true,
          consistent: true
        },
        ObjectPattern: { "multiline": true },
        ImportDeclaration: {
          minProperties: 8,
          multiline: true,
          consistent: true
        }
      }
    ],
    'object-curly-spacing': ['error', 'always'],
    'one-var': ['error', 'never'],
    'one-var-declaration-per-line': ['error', 'always'],
    'operator-assignment': ['error', 'always'],
    'operator-linebreak': [
      'error',
      'before',
      {
        overrides: {
          '=': 'none',
          '&&': 'ignore'
        }
      }
    ],
    'padded-blocks': [
      'error',
      {
        blocks: 'never',
        classes: 'never',
        switches: 'never'
      }
    ],
    'padding-line-between-statements': 'off',
    'quote-props': [
      'error',
      'as-needed',
      {
        keywords: false,
        unnecessary: true,
        numbers: false
      }
    ],
    'quotes': [
      'error',
      'single',
      {
        avoidEscape: true
      }
    ],
    'semi': ['error', 'always'],
    'semi-spacing': [
      'error',
      {
        before: false,
        after: true
      }
    ],
    'semi-style': ['error', 'last'],
    'sort-vars': 'off',
    'space-before-blocks': 'error',
    'space-before-function-paren': [
      'error',
      {
        anonymous: 'always',
        named: 'never',
        asyncArrow: 'always'
      }
    ],
    'space-in-parens': ['error', 'never'],
    'space-infix-ops': 'error',
    'space-unary-ops': [
      'error',
      {
        words: true,
        nonwords: false
      }
    ],
    'spaced-comment': [
      'error',
      'always',
      {
        line: {
          exceptions: ['-', '+'],
          markers: ['=', '!']
        },
        block: {
          exceptions: ['-', '+'],
          markers: ['=', '!'],
          balanced: true
        }
      }
    ],
    'switch-colon-spacing': [
      'error',
      {
        after: true,
        before: false
      }
    ],
    'template-tag-spacing': ['error', 'never'],
    'unicode-bom': ['error', 'never'],
    'wrap-regex': 'off',
    'arrow-body-style': [
      'error',
      'as-needed',
      {
        requireReturnForObjectLiteral: false
      }
    ],
    'arrow-parens': ['error', 'always'],
    'arrow-spacing': [
      'error',
      {
        before: true,
        after: true
      }
    ],
    'generator-star-spacing': [
      'error',
      {
        before: false,
        after: true
      }
    ],
    'no-confusing-arrow': [
      'error',
      {
        allowParens: true
      }
    ],
    'no-useless-computed-key': 'error',
    'no-useless-rename': [
      'error',
      {
        ignoreDestructuring: false,
        ignoreImport: false,
        ignoreExport: false
      }
    ],
    'no-var': 'error',
    'object-shorthand': [
      'error',
      'always',
      {
        ignoreConstructors: false,
        avoidQuotes: true
      }
    ],
    'prefer-arrow-callback': [
      'error',
      {
        allowNamedFunctions: false,
        allowUnboundThis: true
      }
    ],
    'prefer-const': [
      'error',
      {
        destructuring: 'any',
        ignoreReadBeforeAssign: true
      }
    ],
    'prefer-numeric-literals': 'error',
    'prefer-spread': 'error',
    'prefer-template': 'error',
    'rest-spread-spacing': ['error', 'never'],
    'sort-imports': [
      'off',
      {
        ignoreCase: false,
        ignoreMemberSort: false,
        memberSyntaxSortOrder: ['none', 'all', 'multiple', 'single']
      }
    ],
    'template-curly-spacing': 'error',
    'yield-star-spacing': ['error', 'after'],
    'react/no-unknown-property': 'error',
    'react/self-closing-comp': 'error',
    'react/sort-prop-types': [
      'off',
      {
        ignoreCase: true,
        callbacksLast: false,
        requiredFirst: false,
        sortShapeProp: true
      }
    ],
    'react/jsx-boolean-value': ['error', 'never'],
    'react/jsx-closing-bracket-location': ['error', 'line-aligned'],
    'react/jsx-closing-tag-location': 'error',
    'react/jsx-curly-spacing': [
      'error',
      'never',
      {
        allowMultiline: true
      }
    ],
    'react/jsx-equals-spacing': ['error', 'never'],
    'react/jsx-first-prop-new-line': ['error', 'multiline-multiprop'],
    'react/jsx-indent': ['error', 2],
    'react/jsx-indent-props': ['error', 2],
    'react/jsx-max-props-per-line': [
      'error',
      {
        maximum: 1,
        when: 'multiline'
      }
    ],
    'react/jsx-props-no-multi-spaces': 'error',
    'react/jsx-sort-props': [
      'off',
      {
        ignoreCase: true,
        callbacksLast: false,
        shorthandFirst: false,
        shorthandLast: false,
        noSortAlphabetically: false,
        reservedFirst: true
      }
    ],
    'react/jsx-space-before-closing': ['off', 'always'],
    'react/jsx-tag-spacing': [
      'error',
      {
        closingSlash: 'never',
        beforeSelfClosing: 'always',
        afterOpening: 'never',
        beforeClosing: 'never'
      }
    ],
    'react/jsx-wrap-multilines': [
      'error',
      {
        declaration: 'parens-new-line',
        assignment: 'parens-new-line',
        return: 'parens-new-line',
        arrow: 'parens-new-line',
        condition: 'parens-new-line',
        logical: 'parens-new-line',
        prop: 'parens-new-line'
      }
    ],
  }
};
