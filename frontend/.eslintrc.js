module.exports = {
  "parser": "@typescript-eslint/parser",
  "parserOptions": {
    "project": "./tsconfig.json"
  },
  "plugins": [
    "@typescript-eslint",
    "react-hooks",
    "import"
  ],
  "extends": [
    "./.eslintrc.format.js"
  ],
  "env": {
    "browser": true,
    "es6": true,
    "node": true,
    "jest": true
  },
  "globals": {
    "NodeJS": "readonly",
    "Entries": "readonly"
  },
  "overrides": [
    {
      "files": [
        "**/*.ts",
        "**/*.tsx"
      ],
      "rules": {
        '@typescript-eslint/no-unused-vars': [
          1,
          {
            args: 'none',
            ignoreRestSiblings: true
          },
        ],
        "@typescript-eslint/no-shadow": ["error"],
        "@typescript-eslint/no-use-before-define": [
          'error',
          {
            functions: true,
            classes: true,
            variables: true
          }
        ],
        'class-methods-use-this': [
          'error',
          {
            exceptMethods: [
              'render',
              'getInitialState',
              'getDefaultProps',
              'getChildContext',
              'componentWillMount',
              'UNSAFE_componentWillMount',
              'componentDidMount',
              'componentWillReceiveProps',
              'UNSAFE_componentWillReceiveProps',
              'shouldComponentUpdate',
              'componentWillUpdate',
              'UNSAFE_componentWillUpdate',
              'componentDidUpdate',
              'componentWillUnmount',
              'componentDidCatch',
              'getSnapshotBeforeUpdate'
            ]
          }
        ],
        'react/display-name': [
          'off',
          {
            ignoreTranspilerName: false
          }
        ],
        'react/forbid-prop-types': [
          'error',
          {
            forbid: ['any', 'array', 'object'
            ],
            checkContextTypes: true,
            checkChildContextTypes: true
          }
        ],
        'react/forbid-dom-props': ['off'
        ],
        'react/jsx-handler-names': [
          'off',
          {
            eventHandlerPrefix: 'handle',
            eventHandlerPropPrefix: 'on'
          }
        ],
        'react/jsx-key': 1,
        'react/jsx-no-bind': [
          'error',
          {
            ignoreRefs: true,
            allowArrowFunctions: true,
            allowBind: false
          }
        ],
        'react/jsx-no-duplicate-props': [
          'error',
          {
            ignoreCase: true
          }
        ],
        'react/jsx-no-literals': [
          'off',
          {
            noStrings: true
          }
        ],
        'react/jsx-no-undef': 'error',
        'react/jsx-pascal-case': [
          'error',
          {
            allowAllCaps: true
          }
        ],
        'react/jsx-sort-prop-types': 'off',
        'react/jsx-sort-default-props': [
          'off',
          {
            ignoreCase: true
          }
        ],
        'react/jsx-uses-react': ['error'
        ],
        'react/jsx-uses-vars': 'error',
        'react/no-deprecated': ['error'
        ],
        'react/no-did-mount-set-state': 'off',
        'react/no-did-update-set-state': 'off',
        'react/no-will-update-set-state': 'error',
        'react/no-direct-mutation-state': 'off',
        'react/no-is-mounted': 'error',
        'react/no-multi-comp': [
          'error',
          {
            ignoreStateless: true
          }
        ],
        'react/no-set-state': 'off',
        'react/no-string-refs': 'error',
        'react/prefer-es6-class': ['error', 'always'
        ],
        'react/prop-types': "off",
        'react/react-in-jsx-scope': 'error',
        'react/require-render-return': 'error',
        'react/sort-comp': [
          'error',
          {
            order: [
              'static-methods',
              'instance-themes',
              'lifecycle',
              '/^on.+$/',
              'getters',
              'setters',
              '/^(get|set)(?!(InitialState$|DefaultProps$|ChildContext$)).+$/',
              'instance-methods',
              'everything-else',
              'rendering'
            ],
            groups: {
              lifecycle: [
                'displayName',
                'propTypes',
                'contextTypes',
                'childContextTypes',
                'mixins',
                'statics',
                'defaultProps',
                'constructor',
                'getDefaultProps',
                'getInitialState',
                'state',
                'getChildContext',
                'componentWillMount',
                'componentDidMount',
                'componentWillReceiveProps',
                'shouldComponentUpdate',
                'componentWillUpdate',
                'componentDidUpdate',
                'componentWillUnmount'
              ],
              rendering: ['/^render.+$/', 'render'
              ]
            }
          }
        ],
        'react/jsx-no-target-blank': [
          'error',
          {
            enforceDynamicLinks: 'always'
          }
        ],
        'react/jsx-filename-extension': [
          'error',
          {
            extensions: ['.jsx', '.tsx']
          }
        ],
        'react/jsx-no-comment-textnodes': 'error',
        'react/no-render-return-value': 'error',
        'react/require-optimization': ['off'
        ],
        'react/no-find-dom-node': 'error',
        'react/forbid-component-props': ['off'
        ],
        'react/forbid-elements': ['off'
        ],
        'react/no-danger-with-children': 'error',
        'react/no-unused-prop-types': [
          'error',
          {
            skipShapeProps: true
          }
        ],
        'react/style-prop-object': 'error',
        'react/no-unescaped-entities': 'error',
        'react/no-children-prop': 'error',
        'react/no-array-index-key': 'error',
        'react/require-default-props': [
          'error',
          {
            forbidDefaultForRequired: true
          }
        ],
        'react/forbid-foreign-prop-types': [
          'warn',
          {
            allowInPropTypes: true
          }
        ],
        'react/void-dom-elements-no-children': 'error',
        'react/default-props-match-prop-types': [
          'error',
          {
            allowRequiredDefaults: false
          }
        ],
        'react/no-redundant-should-component-update': 'error',
        'react/no-unused-state': 'error',
        "react-hooks/rules-of-hooks": "error",
        "react-hooks/exhaustive-deps": "warn",
        'react/boolean-prop-naming': [
          'off',
          {
            propTypeNames: ['bool', 'mutuallyExclusiveTrueProps'
            ],
            rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+'
          }
        ],
        'react/no-typos': 'error',
        'react/jsx-curly-brace-presence': [
          'error',
          {
            props: 'never',
            children: 'never'
          }
        ],
        'react/jsx-one-expression-per-line': 0,
        'react/destructuring-assignment': ['error', 'always'
        ],
        'react/no-access-state-in-setstate': 'error',
        'react/jsx-child-element-spacing': 'off',
        'react/no-this-in-sfc': 'error',
        'react/jsx-max-depth': 'off',
        'accessor-pairs': 'off',
        'array-callback-return': [
          'error',
          {
            allowImplicit: true
          }
        ],
        'block-scoped-var': 'error',
        'consistent-return': 'error',
        'default-case': [
          'error',
          {
            commentPattern: '^no default$'
          }
        ],
        'guard-for-in': 'error',
        'no-alert': 'warn',
        'no-caller': 'error',
        'no-case-declarations': 'error',
        'no-div-regex': 'off',
        'no-empty-function': [
          'error',
          {
            allow: ['arrowFunctions', 'functions', 'methods'
            ]
          }
        ],
        'no-empty-pattern': 'error',
        'no-eq-null': 'off',
        'no-eval': 'error',
        'no-extend-native': 'error',
        'no-fallthrough': 'error',
        'no-global-assign': ['error'
        ],
        'no-native-reassign': 'off',
        'no-implicit-globals': 'off',
        'no-implied-eval': 'error',
        'no-invalid-this': 'off',
        'no-iterator': 'error',
        'no-labels': [
          'error',
          {
            allowLoop: false,
            allowSwitch: false
          }
        ],
        'no-lone-blocks': 'error',
        'no-loop-func': 'error',
        'no-magic-numbers': [
          'off',
          {
            ignoreArrayIndexes: true,
            enforceConst: true,
            detectObjects: false
          }
        ],
        'no-multi-str': 'error',
        'no-new': 'error',
        'no-new-func': 'error',
        'no-new-wrappers': 'error',
        'no-octal': 'error',
        'no-octal-escape': 'error',
        'no-param-reassign': [
          'error',
          {
            props: true,
            ignorePropertyModificationsFor: [
              'acc',
              'accumulator',
              'e',
              'ctx',
              'req',
              'request',
              'res',
              'response',
              '$scope',
              'state'
            ]
          }
        ],
        'no-proto': 'error',
        'no-redeclare': 'error',
        'no-restricted-properties': [
          'error',
          {
            object: 'arguments',
            property: 'callee',
            message: 'arguments.callee is deprecated'
          },
          {
            object: 'global',
            property: 'isFinite',
            message: 'Please use Number.isFinite instead'
          },
          {
            object: 'self',
            property: 'isFinite',
            message: 'Please use Number.isFinite instead'
          },
          {
            object: 'window',
            property: 'isFinite',
            message: 'Please use Number.isFinite instead'
          },
          {
            object: 'global',
            property: 'isNaN',
            message: 'Please use Number.isNaN instead'
          },
          {
            object: 'self',
            property: 'isNaN',
            message: 'Please use Number.isNaN instead'
          },
          {
            object: 'window',
            property: 'isNaN',
            message: 'Please use Number.isNaN instead'
          },
          {
            property: '__defineGetter__',
            message: 'Please use Object.defineProperty instead.'
          },
          {
            property: '__defineSetter__',
            message: 'Please use Object.defineProperty instead.'
          },
          {
            object: 'Math',
            property: 'pow',
            message: 'Use the exponentiation operator (**) instead.'
          }
        ],
        'no-return-assign': ['error', 'always'
        ],
        'no-return-await': 'error',
        'no-script-url': 'error',
        'no-self-assign': 'error',
        'no-self-compare': 'error',
        'no-sequences': 'error',
        'no-throw-literal': 'error',
        'no-unmodified-loop-condition': 'off',
        'no-unused-expressions': [
          'error',
          {
            allowShortCircuit: true,
            allowTernary: true,
            allowTaggedTemplates: false
          }
        ],
        'no-useless-call': 'off',
        'no-useless-concat': 'error',
        'no-useless-escape': 'error',
        'no-void': 'error',
        'no-warning-comments': [
          'off',
          {
            terms: ['todo', 'fixme', 'xxx'
            ],
            location: 'start'
          }
        ],
        'no-with': 'error',
        'prefer-promise-reject-errors': [
          'error',
          {
            allowEmptyReject: true
          }
        ],
        radix: 'error',
        'require-await': 'off',
        'vars-on-top': 'error',
        'for-direction': 'error',
        'getter-return': [
          'error',
          {
            allowImplicit: true
          }
        ],
        'no-await-in-loop': 'error',
        'no-compare-neg-zero': 'error',
        'no-cond-assign': ['error', 'always'
        ],
        'no-console': 'warn',
        'no-constant-condition': 0,
        'no-control-regex': 'error',
        'no-debugger': 'warn',
        'no-dupe-args': 'error',
        'no-dupe-keys': 'error',
        'no-duplicate-case': 'error',
        'no-empty': 'error',
        'no-empty-character-class': 'error',
        'no-ex-assign': 'error',
        'no-func-assign': 'error',
        'no-inner-declarations': 'error',
        'no-invalid-regexp': 'error',
        'no-irregular-whitespace': 'error',
        'no-obj-calls': 'error',
        'no-prototype-builtins': 'error',
        'no-sparse-arrays': 'error',
        'no-template-curly-in-string': 'error',
        'no-unexpected-multiline': 'error',
        'no-unreachable': 'error',
        'no-unsafe-finally': 'error',
        'no-negated-in-lhs': 'off',
        'use-isnan': 'error',
        'valid-jsdoc': 'off',
        'valid-typeof': [
          'error',
          {
            requireStringLiterals: true
          }
        ],
        'callback-return': 'off',
        'global-require': 'error',
        'handle-callback-err': 'off',
        'no-buffer-constructor': 'error',
        'no-mixed-requires': ['off',
          false
        ],
        'no-new-require': 'error',
        'no-path-concat': 'error',
        'no-process-env': 'off',
        'no-process-exit': 'off',
        'no-restricted-modules': 'off',
        'no-sync': 'off',
        'consistent-this': 'off',
        'func-name-matching': [
          'off',
          'always',
          {
            includeCommonJSModuleExports: false
          }
        ],
        'func-names': 'warn',
        'func-style': ['off', 'expression'
        ],
        'id-blacklist': 'off',
        'id-length': 'off',
        'id-match': 'off',
        'line-comment-position': [
          'off',
          {
            position: 'above',
            ignorePattern: '',
            applyDefaultPatterns: true
          }
        ],
        'lines-around-directive': [
          'error',
          {
            before: 'always',
            after: 'always'
          }
        ],
        'max-depth': ['off',
          4
        ],
        'max-nested-callbacks': 'off',
        'max-params': ['off',
          3
        ],
        'max-statements': ['off',
          10
        ],
        'max-statements-per-line': [
          'off',
          {
            max: 1
          }
        ],
        'multiline-ternary': ['off', 'never'
        ],
        'new-cap': [
          'error',
          {
            newIsCap: true,
            newIsCapExceptions: [],
            capIsNew: false,
            capIsNewExceptions: ['Immutable.Map', 'Immutable.Set', 'Immutable.List'
            ]
          }
        ],
        'newline-after-var': 'off',
        'newline-before-return': 'off',
        'no-array-constructor': 'error',
        'no-bitwise': 'error',
        'no-continue': 'error',
        'no-inline-comments': 'off',
        'no-mixed-operators': [
          'error',
          {
            groups: [
              ['%', '**'
              ],
              ['%', '+'
              ],
              ['%', '-'
              ],
              ['%', '*'
              ],
              ['%', '/'
              ],
              ['**', '+'
              ],
              ['**', '-'
              ],
              ['**', '*'
              ],
              ['**', '/'
              ],
              ['&', '|', '^', '~', '<<', '>>', '>>>'
              ],
              ['==', '!=', '===', '!==', '>', '>=', '<', '<='
              ],
              ['&&', '||'
              ],
              ['in', 'instanceof'
              ]
            ],
            allowSamePrecedence: false
          }
        ],
        'no-mixed-spaces-and-tabs': 'error',
        'no-multi-assign': ['error'
        ],
        'no-negated-condition': 'off',
        'no-nested-ternary': 'off',
        'no-new-object': 'error',
        'no-plusplus': 'error',
        'no-restricted-syntax': [
          'error',
          {
            selector: 'ForInStatement',
            message:
              'for..in loops iterate over the entire prototype chain, which is virtually never what you want. Use Object.{keys,values,entries}, and iterate over the resulting array.'
          },
          {
            selector: 'ForOfStatement',
            message:
              'iterators/generators require regenerator-runtime, which is too heavyweight for this guide to allow them. Separately, loops should be avoided in favor of array iterations.'
          },
          {
            selector: 'LabeledStatement',
            message: 'Labels are a form of GOTO; using them makes code confusing and hard to maintain and understand.'
          },
          {
            selector: 'WithStatement',
            message: '`with` is disallowed in strict mode because it makes code impossible to predict and optimize.'
          }
        ],
        'no-spaced-func': 'error',
        'no-tabs': 'error',
        'no-ternary': 'off',
        'no-underscore-dangle': [
          0,
          {
            allowAfterThis: false,
            allowAfterSuper: false,
            enforceInMethodNames: false
          }
        ],
        'require-jsdoc': 'off',
        'sort-keys': [
          'off',
          'asc',
          {
            caseSensitive: false,
            natural: true
          }
        ],
        'init-declarations': 'off',
        'no-catch-shadow': 'off',
        'no-delete-var': 'error',
        'no-label-var': 'error',
        'no-restricted-globals': ['error', 'isFinite', 'isNaN'
        ],
        'no-shadow': 'off',
        'no-shadow-restricted-names': 'error',
        'no-undef': 'error',
        'no-undefined': 'off',
        'no-unused-vars': 0,
        'no-use-before-define': "off",
        'constructor-super': 'error',
        'no-class-assign': 'error',
        'no-const-assign': 'error',
        'no-dupe-class-members': 'error',
        'no-duplicate-imports': 'off',
        'no-new-symbol': 'error',
        'no-restricted-imports': ['off'
        ],
        'no-this-before-super': 'error',
        'no-useless-constructor': 'error',
        'prefer-destructuring': [
          'error',
          {
            VariableDeclarator: {
              array: false,
              object: true
            },
            AssignmentExpression: {
              array: true,
              object: false
            }
          },
          {
            enforceForRenamedProperties: false
          }
        ],
        'prefer-reflect': 'off',
        'prefer-rest-params': 'error',
        'require-yield': 'error',
        'symbol-description': 'error',
        'import/no-unresolved': [
          'error',
          {
            commonjs: true,
            caseSensitive: true
          }
        ],
        'import/named': 'error',
        'import/default': 'off',
        'import/namespace': 'off',
        'import/export': 'error',
        'import/no-named-as-default': 'error',
        'import/no-named-as-default-member': 'error',
        'import/no-deprecated': 'off',
        'import/no-extraneous-dependencies': 0,
        'import/no-mutable-exports': 'error',
        'import/no-commonjs': 'off',
        'import/no-amd': 'error',
        'import/no-nodejs-modules': 'off',
        'import/first': 'error',
        'import/imports-first': 'off',
        'import/no-duplicates': 'error',
        'import/no-namespace': 'off',
        'import/extensions': [
          'error',
          'ignorePackages',
          {
            js: 'never',
            ts: 'never',
            tsx: 'never',
            mjs: 'never',
            jsx: 'never'
          }
        ],
        'import/order': [
          'error',
          {
            "newlines-between": "always",
            pathGroups: [
              {
                "pattern": "{@*,@*/**}",
                "group": "external",
                "position": "after"
              }
            ],
            groups: [
              ["builtin", "external", "internal"],
            ]
          }
        ],
        'import/newline-after-import': 'error',
        'import/no-restricted-paths': 'off',
        'import/max-dependencies': [
          'off',
          {
            max: 10
          }
        ],
        'import/no-absolute-path': 'error',
        'import/no-dynamic-require': 'error',
        'import/no-internal-modules': ['off'
        ],
        'import/unambiguous': 'off',
        'import/no-webpack-loader-syntax': 'error',
        'import/no-unassigned-import': 'off',
        'import/no-named-default': 'error',
        'import/no-anonymous-default-export': [
          'off',
          {
            allowArray: false,
            allowArrowFunction: false,
            allowAnonymousClass: false,
            allowAnonymousFunction: false,
            allowLiteral: false,
            allowObject: false
          }
        ],
        'import/exports-last': 'off',
        'import/group-exports': 'off',
        'import/no-default-export': 'off',
        'import/no-self-import': 'error',
        'import/no-useless-path-segments': 'error',
        'import/dynamic-import-chunkname': [
          'off',
          {
            importFunctions: [],
            webpackChunknameFormat: '[0-9a-zA-Z-_/.]+'
          }
        ]
      }
    }
  ],
}
