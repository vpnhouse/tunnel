import { FC, useCallback, FormEvent, useState, ChangeEvent } from 'react';
import { Navigate } from 'react-router-dom';
import { useUnit } from 'effector-react';
import Box from '@mui/material/Box';
import { useTheme } from '@mui/material/styles';

import { MAIN_ROUTE } from '@constants/routes';
import { TextField, Button } from '@common/ui-kit/components';
import { VisibilityAdornment } from '@common/components';
import { $authStore, loginFx } from '@root/store/auth';
import { AuthDataType } from '@root/store/auth/types';
import logo from '@common/assets/logo.png';

const Auth: FC = () => {
  const theme = useTheme();
  const isAuthenticated = useUnit($authStore);
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

  if (isAuthenticated) {
    return <Navigate to={MAIN_ROUTE} replace />;
  }

  return (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        backgroundColor: theme.palette.background.default
      }}
    >
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center'
        }}
      >
        <Box
          component="img"
          src={logo}
          alt="logo"
          sx={{
            width: 130,
            height: 32,
            marginBottom: 4
          }}
        />

        <Box
          component="form"
          onSubmit={formSubmitHandler}
          sx={{ width: '320px' }}
        >
          <TextField
            sx={{ marginBottom: 1.5, marginTop: 0 }}
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
        </Box>
      </Box>
    </Box>
  );
};

export default Auth;
