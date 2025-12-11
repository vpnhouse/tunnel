import { ChangeEvent, FC, useId, useMemo } from 'react';
import Box from '@mui/material/Box';
import { styled, useTheme } from '@mui/material/styles';

interface Props {
  options: [string, string];
  selected: string;
  onChange: (event: ChangeEvent<HTMLElement>) => void;
  labels?: [string, string];
  name?: string;
  /** Accessible label for screen readers */
  ariaLabel?: string;
}

const SwitcherRoot = styled(Box)({
  display: 'flex',
  fontSize: '16px',
  width: 260,
  marginBottom: 12,
  backgroundColor: '#2B3142',
  borderRadius: 8,
  padding: 4,
  '&>input': {
    display: 'none'
  }
});

const SwitcherOption = styled('label')(({ theme }) => ({
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  padding: '8px 12px',
  height: 24,
  cursor: 'pointer',
  position: 'relative',
  borderRadius: 8,
  flexGrow: 1,
  textTransform: 'capitalize',
  fontWeight: 400,
  fontSize: 16,
  lineHeight: '14px',
  transition: 'background-color 0.5s ease',
  '&:last-child': {
    marginLeft: '-5px'
  }
}));

const Switcher: FC<Props> = ({ options, selected, onChange, name, labels, ariaLabel }) => {
  const theme = useTheme();
  const reactId = useId();
  const [firstOption, secondOption] = options;
  const switcherName = name ?? 'switcher';

  // Generate truly unique IDs using React's useId + name + random suffix
  const uniqueIds = useMemo(() => {
    const baseId = reactId.replace(/:/g, '-');
    return {
      first: `switcher${baseId}${switcherName}-opt1`,
      second: `switcher${baseId}${switcherName}-opt2`
    };
  }, [reactId, switcherName]);

  // Use labels or options for accessible names
  const firstLabel = labels?.[0] ?? firstOption;
  const secondLabel = labels?.[1] ?? secondOption;

  const getOptionStyles = (isActive: boolean) => ({
    ...(isActive ? {
      backgroundColor: theme.palette.primary.main,
      cursor: 'unset',
      zIndex: 2
    } : {
      '&:hover': {
        backgroundColor: '#3B3F63'
      },
      '&>label': {
        cursor: 'pointer'
      }
    })
  });

  return (
    <SwitcherRoot
      role="radiogroup"
      aria-label={ariaLabel ?? `${switcherName} options`}
    >
      <input
        type="radio"
        value={firstOption}
        name={switcherName}
        id={uniqueIds.first}
        checked={selected === firstOption}
        onChange={onChange}
        aria-label={firstLabel}
        title={firstLabel}
      />
      <SwitcherOption
        sx={getOptionStyles(selected === firstOption)}
        htmlFor={uniqueIds.first}
      >
        {firstLabel}
      </SwitcherOption>

      <input
        type="radio"
        value={secondOption}
        name={switcherName}
        id={uniqueIds.second}
        checked={selected === secondOption}
        onChange={onChange}
        aria-label={secondLabel}
        title={secondLabel}
      />
      <SwitcherOption
        sx={getOptionStyles(selected === secondOption)}
        htmlFor={uniqueIds.second}
      >
        {secondLabel}
      </SwitcherOption>
    </SwitcherRoot>
  );
};

export default Switcher;
