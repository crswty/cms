import {defaultTheme} from "react-admin";
import {ThemeOptions} from "@mui/material";

export const CustomTheme = {
    ...defaultTheme,
    components: {
        MuiTextField: {
            defaultProps: {
                variant: "outlined"
            }
        }
    },
    sidebar: {
        width: 150,
        closedWidth: 150
    }
} as ThemeOptions