package io.manun.camli;

import android.content.Context;
import android.content.Intent;
import android.os.Bundle;
import android.renderscript.ScriptGroup;
import android.text.InputType;
import android.util.Log;
import android.widget.EditText;

import androidx.annotation.NonNull;
import androidx.appcompat.app.ActionBar;
import androidx.appcompat.app.AppCompatActivity;
import androidx.preference.EditTextPreference;
import androidx.preference.Preference;
import androidx.preference.PreferenceFragmentCompat;

public class SettingsActivity extends AppCompatActivity {

    private static final String TAG = SettingsActivity.class.getName();

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.settings_activity);
        if (savedInstanceState == null) {
            getSupportFragmentManager()
                    .beginTransaction()
                    .replace(R.id.settings, new SettingsFragment())
                    .commit();
        }
        ActionBar actionBar = getSupportActionBar();
        if (actionBar != null) {
            actionBar.setDisplayHomeAsUpEnabled(true);
        }
    }

    public static class SettingsFragment extends PreferenceFragmentCompat {
        private EditTextPreference hostPref;
        private EditTextPreference passwordPref;

        @Override
        public void onCreatePreferences(Bundle savedInstanceState, String rootKey) {
            setPreferencesFromResource(R.xml.root_preferences, rootKey);
            hostPref = findPreference(Preferences.HOST);
            passwordPref = findPreference(Preferences.PASSWORD);

            if (passwordPref != null) {
                passwordPref.setOnBindEditTextListener(editText ->
                        editText.setInputType(
                                InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_VARIATION_PASSWORD
                        ));
            }

            Preference.OnPreferenceChangeListener onChange = (preference, newValue) -> {
                final String key = preference.getKey();
                Log.v(TAG, "Preference change for: " + key);
                String newStr = (newValue instanceof String) ? (String) newValue : null;
                if (preference == hostPref) {
                    updateHostSummary(newStr);
                }
                if (preference == passwordPref) {
                    updatePasswordSummary(newStr);
                }
                return true;
            };

            hostPref.setOnPreferenceChangeListener(onChange);
            passwordPref.setOnPreferenceChangeListener(onChange);
        }

        @Override
        public void onResume() {
            super.onResume();
            updatePreferenceSummaries();
        }

        private void updatePreferenceSummaries() {
            updateHostSummary(hostPref.getText());
            updatePasswordSummary(passwordPref.getText());
        }

        private void updateHostSummary(String value) {
            if (value != null && value.length() > 0) {
                hostPref.setSummary(value);
            } else {
                hostPref.setSummary(getString(R.string.settings_host_summary));
            }
        }

        private void updatePasswordSummary(String value) {
            if (value != null && value.length() > 0) {
                passwordPref.setSummary("*********");
            } else {
                passwordPref.setSummary("<unset>");
            }
        }
    }

    static void show(Context context) {
        final Intent intent = new Intent(context, SettingsActivity.class);
        context.startActivity(intent);
    }

}