// IUploadService.aidl
package io.manun.camli;

import io.manun.camli.IStatusCallback;
import android.os.ParcelFileDescriptor;

// Declare any non-default types here with import statements

interface IUploadService {
  void registerCallback(IStatusCallback ob);
  void unregisterCallback(IStatusCallback ob);

  boolean isUploading();

  void stop();
  void start();

  // Returns false if server not configured.
  boolean addFile(in ParcelFileDescriptor pfd);

}