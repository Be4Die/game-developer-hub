<?php
class AdminerPrefillPassword {
    function loginFormField($name, $heading, $value) {
        if ($name === "password" && !empty($_GET["password"])) {
            return $heading . "<input type=\"password\" name=\"auth[password]\" value=\"" . htmlspecialchars($_GET["password"], ENT_QUOTES) . "\" autocomplete=\"current-password\">";
        }
    }
}
return new AdminerPrefillPassword;
