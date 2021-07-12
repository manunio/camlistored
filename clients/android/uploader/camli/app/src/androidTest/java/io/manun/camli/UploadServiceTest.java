package io.manun.camli;

import android.content.Intent;
import android.net.Uri;
import android.os.IBinder;
import android.os.ParcelFileDescriptor;

import androidx.test.core.app.ApplicationProvider;
import androidx.test.ext.junit.runners.AndroidJUnit4;
import androidx.test.rule.ServiceTestRule;

import org.junit.Assert;
import org.junit.Rule;
import org.junit.Test;
import org.junit.runner.RunWith;

import java.util.concurrent.TimeoutException;

import static org.junit.Assert.*;

@RunWith(AndroidJUnit4.class)
public class UploadServiceTest {

    @Rule
    public final ServiceTestRule mServiceRule = new ServiceTestRule();

    @Test
    public void testService() throws TimeoutException {
        // create a new service intent
        Intent intent = new Intent(ApplicationProvider.getApplicationContext(), UploadService.class);

        // Bind service and grab reference to the binder.
        IBinder binder = mServiceRule.bindService(intent);

        // Get the reference to the service,
        // or you can call public method directly.

        UploadService service = ((UploadService.UploadBinder) binder).getService();

        assertNull(service.getFileDescriptor(null));
    }
}