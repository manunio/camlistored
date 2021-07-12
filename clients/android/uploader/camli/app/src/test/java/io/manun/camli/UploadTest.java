package io.manun.camli;

import android.util.Log;

import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.Mockito;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.modules.junit4.PowerMockRunner;

@RunWith(PowerMockRunner.class)
@PrepareForTest({Upload.class, Log.class})
public class UploadTest {

    @Before
    public void beforeTest() {
        PowerMockito.mockStatic(Log.class);
    }

    @Test
    public void run() {
        UploadService uploadServiceMock = Mockito.mock(UploadService.class);
        HostPort hp = new HostPort("localhost:3179");
        Upload upload = new Upload(uploadServiceMock, hp, "foo");
        upload.run();
    }

}