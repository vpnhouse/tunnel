import React, { FC, useCallback, FormEvent, useState, ChangeEvent } from 'react';
import { Redirect } from 'react-router-dom';
import { useStore } from 'effector-react';

import { MAIN_ROUTE } from '@constants/routes';
import { TextField, Button } from '@common/ui-kit/components';
import { VisibilityAdornment } from '@common/components';
import { $authStore, loginFx } from '@root/store/auth';
import { AuthDataType } from '@root/store/auth/types';
import logo from '@common/assets/logo.png';

import useStyles from './index.styles';

const Auth: FC = () => {
  const classes = useStyles();
  const isAuthenticated = useStore($authStore);
  const [authData, setAuthData] = useState<AuthDataType>({
    password: ''
  });
  const [showPassword, setShowPassword] = useState(false);

  const [error, setError] = useState<AuthDataType>({
    password: ''
  });

  const authDataChangeHandler = useCallback((e: ChangeEvent<HTMLInputElement>) => {
    const { value, name } = e.target;
    setError((prevError) => ({
      ...prevError,
      [name]: ''
    }));
    setAuthData((prevAuthData) => ({
      ...prevAuthData,
      [name]: value
    }));
  }, []);

  const toggleShowPasswordHandler = useCallback(() =>
    setShowPassword((prevState) => !prevState), []);

  const validate = useCallback(() => {
    const { password } = authData;

    setError({
      ...error,
      password: password ? '' : 'The password is required'
    });

    return !!password;
  }, [authData, error]);

  const formSubmitHandler = useCallback((e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation();
    e.preventDefault();

    if (!validate()) return;

    loginFx(authData);
  }, [authData, validate]);

  return (
    <>
      {isAuthenticated
        ? <Redirect to={MAIN_ROUTE} />
        : (
          <div className={classes.root}>
            <div className={classes.enterGroup}>
              <img className={classes.logo} src={logo} alt="logo" />

              <form onSubmit={formSubmitHandler} className={classes.form}>
                <TextField
                  className={classes.passwordInput}
                  fullWidth
                  variant="outlined"
                  label="Password"
                  type={showPassword ? 'text' : 'password'}
                  name="password"
                  value={authData.password}
                  helperText={error.password}
                  error={!!error.password}
                  onChange={authDataChangeHandler}
                  endAdornment={(
                    <VisibilityAdornment
                      showPassword={showPassword}
                      toggleShowPasswordHandler={toggleShowPasswordHandler}
                    />
                  )}
                />
                <Button
                  type="submit"
                  variant="contained"
                  color="primary"
                  fullWidth
                  disabled={!authData.password}
                >
                  Log in
                </Button>
              </form>
            </div>
          </div>
        )}
    </>
  );
};

export default Auth;
